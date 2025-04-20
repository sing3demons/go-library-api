package kp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// type httpService struct {
// 	requestAttributes []RequestAttributes
// 	detailLog         logger.DetailLog
// 	summaryLog        logger.SummaryLog
// }

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

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		responses []ApiResponse
	)

	concurrencyLimit := 5 // 👈 limit how many goroutines run at once
	semaphore := make(chan struct{}, concurrencyLimit)
	responseChan := make(chan ApiResponse, len(requestAttributes))

	for _, attr := range requestAttributes {
		wg.Add(1)
		attrCopy := attr // avoid loop variable capture

		go func(attr RequestAttributes) {
			defer wg.Done()
			semaphore <- struct{}{}        // acquire semaphore
			defer func() { <-semaphore }() // release semaphore

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

			// Path param substitution
			for key, value := range attr.Params {
				if strings.Contains(attr.URL, "{"+key+"}") {
					attr.URL = strings.ReplaceAll(attr.URL, "{"+key+"}", value)
				} else if strings.Contains(attr.URL, ":"+key) {
					attr.URL = strings.ReplaceAll(attr.URL, ":"+key, value)
				} else {
					attr.URL = strings.ReplaceAll(attr.URL, key, value)
				}
			}
			processLog.Url = attr.URL

			// Query params
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
				responseChan <- ApiResponse{
					Status: 0,
					attr:   attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
					Err:    err,
				}
				return
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				responseChan <- ApiResponse{
					Status:     resp.StatusCode,
					attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
					StatusText: resp.Status,
					Err:        err,
				}
				return
			}

			var body interface{}
			if err := json.Unmarshal(bodyBytes, &body); err != nil {
				responseChan <- ApiResponse{
					Status:     resp.StatusCode,
					attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
					RawBody:    string(bodyBytes),
					StatusText: resp.Status,
					Err:        err,
				}
				return
			}

			responseChan <- ApiResponse{
				Status:     resp.StatusCode,
				attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
				Header:     resp.Header,
				Body:       body,
				RawBody:    string(bodyBytes),
				StatusText: resp.Status,
			}
		}(attrCopy)
	}

	wg.Wait()
	close(responseChan)

	for response := range responseChan {
		service := response.attr.Service
		command := response.attr.Command
		invoke := response.attr.Invoke
		method := response.attr.Method

		resultCode := fmt.Sprintf("%d", response.Status)
		summaryLog.AddSuccess(service, command, resultCode, response.StatusText)
		detailLog.AddInputResponse(service, command, invoke, nil, response, "http", string(method))

		mu.Lock()
		responses = append(responses, response)
		mu.Unlock()
	}

	if len(responses) == 1 {
		return responses[0].Body, nil
	}
	return responses, nil
}

// func RequestHttp(optionAttributes OptionAttributes, detailLog logger.DetailLog, summaryLog logger.SummaryLog) (any, error) {
// 	var requestAttributes []RequestAttributes
// 	switch attr := optionAttributes.(type) {
// 	case []RequestAttributes:
// 		requestAttributes = attr
// 	case RequestAttributes:
// 		requestAttributes = []RequestAttributes{attr}
// 	default:
// 		return nil, errors.New("invalid optionAttributes type")
// 	}

// 	var responses []ApiResponse

// 	for _, attr := range requestAttributes {
// 		processLog := ProcessLog{
// 			Header:      attr.Headers,
// 			Url:         attr.URL,
// 			QueryString: attr.Query,
// 			Body:        attr.Body,
// 			Method:      attr.Method,
// 			RetryCount:  attr.RetryCount,
// 			Timeout:     attr.Timeout,
// 			Auth:        attr.Auth,
// 		}

// 		if len(attr.Params) > 0 {
// 			// Replace URL params
// 			for key, value := range attr.Params {
// 				// startWith "{}"
// 				if strings.Contains(attr.URL, "{"+key+"}") {
// 					attr.URL = strings.ReplaceAll(attr.URL, "{"+key+"}", value)
// 				} else if strings.Contains(attr.URL, ":"+key) {
// 					attr.URL = strings.ReplaceAll(attr.URL, ":"+key, value)
// 				} else {
// 					attr.URL = strings.ReplaceAll(attr.URL, key, value)
// 				}
// 			}
// 			processLog.Url = attr.URL
// 		}

// 		if len(attr.Query) > 0 {
// 			query := url.Values{}
// 			for key, value := range attr.Query {
// 				query.Add(key, value)
// 			}
// 			attr.URL = fmt.Sprintf("%s?%s", attr.URL, query.Encode())
// 			processLog.QueryString = attr.Query
// 		}

// 		detailLog.AddOutputRequest(attr.Service, attr.Command, attr.Invoke, processLog, processLog)

// 		resp, err := SendRequest(RequestAttr{
// 			Method:  string(attr.Method),
// 			URL:     attr.URL,
// 			Headers: attr.Headers,
// 			Body:    attr.Body,
// 			Timeout: attr.Timeout,
// 		})
// 		if err != nil {
// 			responseChan := ApiResponse{
// 				Status:     resp.StatusCode,
// 				attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
// 				Header:     resp.Header,
// 				RawBody:    "",
// 				Body:       nil,
// 				StatusText: resp.Status,
// 				Err:        err,
// 			}
// 			responses = append(responses, responseChan)
// 			continue
// 		}

// 		defer resp.Body.Close()
// 		bodyBytes, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			responseChan := ApiResponse{
// 				Status:     resp.StatusCode,
// 				attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
// 				Header:     nil,
// 				Body:       nil,
// 				StatusText: resp.Status,
// 				Err:        err,
// 			}
// 			responses = append(responses, responseChan)
// 			continue
// 		}
// 		var body interface{}
// 		if err := json.Unmarshal(bodyBytes, &body); err != nil {
// 			responseChan := ApiResponse{
// 				Status:     resp.StatusCode,
// 				attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
// 				Header:     resp.Header,
// 				RawBody:    string(bodyBytes),
// 				Body:       nil,
// 				StatusText: resp.Status,
// 				Err:        err,
// 			}
// 			responses = append(responses, responseChan)
// 			continue
// 		}
// 		responseChan := ApiResponse{
// 			Status:     resp.StatusCode,
// 			attr:       attrDetailLog{Service: attr.Service, Command: attr.Command, Invoke: attr.Invoke, Method: attr.Method},
// 			Header:     resp.Header,
// 			Body:       body,
// 			RawBody:    string(bodyBytes),
// 			StatusText: resp.Status,
// 			Err:        err,
// 		}

// 		responses = append(responses, responseChan)
// 	}
// 	// return service.requestHttp()

// 	for _, response := range responses {
// 		service := response.attr.Service
// 		command := response.attr.Command
// 		invoke := response.attr.Invoke
// 		method := response.attr.Method

// 		resultCode := fmt.Sprintf("%d", response.Status)
// 		summaryLog.AddSuccess(service, command, resultCode, response.StatusText)

// 		detailLog.AddInputResponse(service, command, invoke, nil, response, "http", string(method))
// 	}

// 	if len(responses) == 1 {
// 		return responses[0].Body, nil
// 	}
// 	return responses, nil
// }

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
