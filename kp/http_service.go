package kp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sing3demons/go-library-api/kp/logger"
)

type TMap map[string]string

type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
)

type RequestAttributes struct {
	Headers        TMap
	Method         HTTPMethod
	Params         TMap
	Query          TMap
	Body           TMap
	RetryCondition string
	RetryCount     int
	Timeout        int
	Service        string
	Command        string
	Invoke         string
	URL            string
	Auth           *BasicAuth
	StatusSuccess  []int
}

type OptionAttributes interface{}

type BasicAuth struct {
	Username string
	Password string
}

// attr.Service, attr.Command, attr.Invoke
type attrDetailLog struct {
	Service string
	Command string
	Invoke  string
	Method  HTTPMethod
}

type ApiResponse struct {
	Err        error
	Header     http.Header
	Body       any
	RawBody    string
	Status     int
	StatusText string
	attr       attrDetailLog
}

type httpService struct {
	requestAttributes []RequestAttributes
	detailLog         logger.DetailLog
	summaryLog        logger.SummaryLog
}

func RequestHttp(optionAttributes OptionAttributes, detailLog logger.DetailLog, summaryLog logger.SummaryLog) (any, error) {
	var requestAttributes []RequestAttributes
	switch attr := optionAttributes.(type) {
	case []RequestAttributes:
		requestAttributes = attr
	case RequestAttributes:
		requestAttributes = []RequestAttributes{attr}
	default:
		return nil, errors.New("invalid optionAttributes type")
	}

	// Use a channel to collect responses
	// responseChan := make(chan ApiResponse, len(requestAttributes))
	var responses []ApiResponse

	for _, attr := range requestAttributes {
		processLog := ProcessLog{
			Header:      attr.Headers,
			Url:         attr.URL,
			QueryString: attr.Query,
			Body:        attr.Body,
			Method:      attr.Method,
			RetryCount:  attr.RetryCount,
			Timeout:     attr.Timeout,
			Auth:        attr.Auth,
		}

		if len(attr.Params) > 0 {
			// Replace URL params
			for key, value := range attr.Params {
				// startWith "{}"
				if strings.Contains(attr.URL, "{"+key+"}") {
					attr.URL = strings.ReplaceAll(attr.URL, "{"+key+"}", value)
				} else if strings.Contains(attr.URL, ":"+key) {
					attr.URL = strings.ReplaceAll(attr.URL, ":"+key, value)
				} else {
					attr.URL = strings.ReplaceAll(attr.URL, key, value)
				}
			}
			processLog.Url = attr.URL
		}

		if len(attr.Query) > 0 {
			query := url.Values{}
			for key, value := range attr.Query {
				query.Add(key, value)
			}
			attr.URL = fmt.Sprintf("%s?%s", attr.URL, query.Encode())
			processLog.QueryString = attr.Query
		}

		detailLog.AddOutputRequest(attr.Service, attr.Command, attr.Invoke, processLog, processLog)

		resp, err := SendRequest(RequestAttr{
			Method:  string(attr.Method),
			URL:     attr.URL,
			Headers: attr.Headers,
			Body:    attr.Body,
			Timeout: attr.Timeout,
		})
		if err != nil {
			responseChan := ApiResponse{
				Status:     resp.StatusCode,
				attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
				Header:     resp.Header,
				RawBody:    "",
				Body:       nil,
				StatusText: resp.Status,
				Err:        err,
			}
			responses = append(responses, responseChan)
			continue
		}

		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			responseChan := ApiResponse{
				Status:     resp.StatusCode,
				attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
				Header:     nil,
				Body:       nil,
				StatusText: resp.Status,
				Err:        err,
			}
			responses = append(responses, responseChan)
			continue
		}
		var body interface{}
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			responseChan := ApiResponse{
				Status:     resp.StatusCode,
				attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
				Header:     resp.Header,
				RawBody:    string(bodyBytes),
				Body:       nil,
				StatusText: resp.Status,
				Err:        err,
			}
			responses = append(responses, responseChan)
			continue
		}
		responseChan := ApiResponse{
			Status:     resp.StatusCode,
			attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
			Header:     resp.Header,
			Body:       body,
			RawBody:    string(bodyBytes),
			StatusText: resp.Status,
			Err:        err,
		}

		responses = append(responses, responseChan)
	}
	// return service.requestHttp()

	for _, response := range responses {
		service := response.attr.Service
		command := response.attr.Command
		invoke := response.attr.Invoke
		method := response.attr.Method

		resultCode := fmt.Sprintf("%d", response.Status)
		summaryLog.AddSuccess(service, command, resultCode, response.StatusText)

		detailLog.AddInputResponse(service, command, invoke, nil, response, "http", string(method))
	}

	if len(responses) == 1 {
		return responses[0].Body, nil
	}
	return responses, nil
}

type RequestAttr struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    any
	Timeout int // in seconds
}

func SendRequest(attr RequestAttr) (*http.Response, error) {
	// Encode body if present
	var body io.Reader
	if attr.Body != nil {
		jsonBytes, err := json.Marshal(attr.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonBytes)
	}

	// Create request
	req, err := http.NewRequest(string(attr.Method), attr.URL, body)
	if err != nil {
		return nil, err
	}

	// Set Content-Type if body exists
	if attr.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range attr.Headers {
		req.Header.Set(key, value)
	}

	// Send request
	httpClient := &http.Client{
		Timeout: time.Duration(attr.Timeout) * time.Second,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request: ===============>", err)
		return nil, err
	}

	// Check for non-2xx status codes
	fmt.Println("Response Status Code: ===============>", resp.StatusCode)

	return resp, nil
}

func (svc *httpService) requestHttp() (any, error) {
	var wg sync.WaitGroup

	// Use a channel to collect responses
	responseChan := make(chan ApiResponse, len(svc.requestAttributes))
	semaphore := make(chan struct{}, 100) // limit to 100 goroutines

	for _, attr := range svc.requestAttributes {
		semaphore <- struct{}{}

		config := RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 100 * time.Millisecond,
			MaxDelay:     1 * time.Second,
		}

		wg.Add(1)
		go func(attr RequestAttributes) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			defer wg.Done()
			defer func() { <-semaphore }()

			transport := NewRetryRoundTripper(
				http.DefaultTransport,
				config,
			)

			req, err := svc.createRequest(ctx, attr)
			if err != nil {
				responseChan <- ApiResponse{
					Status:     500,
					attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
					Header:     nil,
					Body:       nil,
					StatusText: err.Error(),
					Err:        err,
				}
				return
			}

			client := &http.Client{
				Timeout:   time.Duration(attr.Timeout) * time.Second,
				Transport: transport,
			}
			response := svc.executeRequest(client, req, attr)

			// mu.Lock()
			// defer mu.Unlock()
			responseChan <- *response
		}(attr)
	}

	fmt.Println("Number of goroutines:", runtime.NumGoroutine())
	wg.Wait()
	close(responseChan)
	svc.detailLog.AutoEnd()

	var responses []any
	for response := range responseChan {
		service := response.attr.Service
		command := response.attr.Command
		invoke := response.attr.Invoke
		method := response.attr.Method

		resultCode := fmt.Sprintf("%d", response.Status)
		svc.summaryLog.AddSuccess(service, command, resultCode, response.StatusText)
		// remove attr from response
		// delete(response.Body, "attr")
		svc.detailLog.AddInputResponse(service, command, invoke, nil, response, "http", string(method))
		responses = append(responses, response.Body)
	}

	if len(responses) == 1 {
		return responses[0], nil
	}
	return responses, nil
}

type ProcessLog struct {
	Header      TMap       `json:"Header"`
	Url         string     `json:"Url"`
	QueryString TMap       `json:"QueryString"`
	Body        TMap       `json:"Body"`
	Method      HTTPMethod `json:"Method"`
	RetryCount  int        `json:"RetryCount,omitempty"`
	Timeout     int        `json:"Timeout,omitempty"`
	Auth        *BasicAuth `json:"Auth,omitempty"`
}

func (svc *httpService) createRequest(ctx context.Context, attr RequestAttributes) (*http.Request, error) {
	processLog := ProcessLog{
		Header:      attr.Headers,
		Url:         attr.URL,
		QueryString: attr.Query,
		Body:        attr.Body,
		Method:      attr.Method,
		RetryCount:  attr.RetryCount,
		Timeout:     attr.Timeout,
		Auth:        attr.Auth,
	}

	if len(attr.Params) > 0 {
		// Replace URL params
		for key, value := range attr.Params {
			// startWith "{}"
			if strings.Contains(attr.URL, "{"+key+"}") {
				attr.URL = strings.ReplaceAll(attr.URL, "{"+key+"}", value)
			} else if strings.Contains(attr.URL, ":"+key) {
				attr.URL = strings.ReplaceAll(attr.URL, ":"+key, value)
			} else {
				attr.URL = strings.ReplaceAll(attr.URL, key, value)
			}
		}
		processLog.Url = attr.URL
	}

	if len(attr.Query) > 0 {
		query := url.Values{}
		for key, value := range attr.Query {
			query.Add(key, value)
		}
		attr.URL = fmt.Sprintf("%s?%s", attr.URL, query.Encode())
		processLog.QueryString = attr.Query
	}
	var body io.Reader
	// var bodyBytes []byte
	if len(attr.Body) > 0 {
		bodyBytes, err := json.Marshal(attr.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(bodyBytes)
		processLog.Body = attr.Body
	}

	svc.detailLog.AddOutputRequest(attr.Service, attr.Command, attr.Invoke, processLog, processLog)

	// Create request
	// req, err := http.NewRequestWithContext(ctx, string(attr.Method), attr.URL, body)
	req, err := http.NewRequest(string(attr.Method), attr.URL, body)
	if err != nil {
		// detailLog.AutoEnd()
		svc.summaryLog.AddError(attr.Service, attr.Command, "500", err.Error())
		// detailLog.AddInputResponse(attr.Service, attr.Command, attr.Invoke, err.Error(), err.Error(), "http", string(attr.Method))
		return nil, err
	}

	// Set headers
	for key, value := range attr.Headers {
		req.Header.Set(key, value)
	}

	// Set BasicAuth
	if attr.Auth != nil {
		req.SetBasicAuth(attr.Auth.Username, attr.Auth.Password)
	}

	return req, nil
}

func (svc *httpService) executeRequest(client *http.Client, req *http.Request, attr RequestAttributes) *ApiResponse {
	apiResponse := &ApiResponse{
		attr: attrDetailLog{
			Service: attr.Service,
			Command: attr.Command,
			Invoke:  attr.Invoke,
			Method:  attr.Method,
		},
	}

	response, err := client.Do(req)
	if err != nil {
		apiResponse.Status = 500
		apiResponse.StatusText = err.Error()
		apiResponse.Err = err
		return apiResponse
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		apiResponse.Status = 500
		apiResponse.StatusText = err.Error()
		apiResponse.Err = err
		return apiResponse
	}

	apiResponse.Header = response.Header
	apiResponse.Status = response.StatusCode
	apiResponse.StatusText = response.Status

	if len(bodyBytes) == 0 {
		fmt.Println("Empty response body")
		apiResponse.Body = nil
		return apiResponse
	}

	var body interface{}
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		// fallback to raw string if not JSON
		apiResponse.Body = string(bodyBytes)
		apiResponse.Err = err
		apiResponse.StatusText = "invalid JSON response"
		return apiResponse
	}

	apiResponse.Body = body
	return apiResponse
}
