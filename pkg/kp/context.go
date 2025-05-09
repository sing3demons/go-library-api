package kp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sing3demons/go-library-api/pkg/kp/logger"
	"go.opentelemetry.io/otel"
)

type IContext interface {
	Context() context.Context
	SetHeader(key, value string)
	GetHeader(key string) string
	// Header() http.Header
	// FullPath() string
	// GetMethod() string
	Next()

	Log() ILogger
	Param(name string) string
	Query(name string) string
	ReadInput(data any) error
	Response(code int, data any) error

	SendMessage(topic string, payload any, opts ...OptionProducerMsg) (RecordMetadata, error)
	CommonLog(cmd, scenario string)
	DetailLog() logger.DetailLog
	SummaryLog() logger.SummaryLog
}

type HandleFunc func(ctx IContext) error

type ServiceHandleFunc HandleFunc

type Middleware func(HandleFunc) HandleFunc

type HttpContext struct {
	ctx         *gin.Context
	cfg         *KafkaConfig
	log         ILogger
	detailLog   logger.DetailLog
	summaryLog  logger.SummaryLog
	baseCommand string
	initInvoke  string
	copyBody    []byte

	// handlers []HandleFunc
	// index    int
}

func newMuxContext(c *gin.Context, cfg *KafkaConfig, log ILogger) IContext {
	ctx := InitSession(c.Request.Context(), log)
	c.Request = c.Request.WithContext(ctx)
	return &HttpContext{ctx: c, cfg: cfg, log: log}
}

func (c *HttpContext) Incoming() logger.InComing {
	var data logger.InComing

	// check method if GET or DELETE not request body
	if c.ctx.Request.Method == "GET" || c.ctx.Request.Method == "DELETE" {
		c.copyBody = nil
	} else {
		// --- Copy and parse body ---
		var body map[string]any
		rawBody, err := io.ReadAll(c.ctx.Request.Body)
		if err == nil && len(rawBody) > 0 {
			// Restore body stream for future reads
			c.ctx.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

			// Try parsing JSON
			if err := json.Unmarshal(rawBody, &body); err == nil && len(body) > 0 {
				data.Body = body
			}
		}
	}

	// --- Copy headers ---
	if headers := c.ctx.Request.Header; len(headers) > 0 {
		headerMap := make(map[string]any, len(headers))
		for key, values := range headers {
			if len(values) > 0 {
				headerMap[key] = values[0]
			}
		}
		data.Header = headerMap
	}

	// --- Copy query string ---
	if query := c.ctx.Request.URL.Query(); len(query) > 0 {
		data.QueryString = query
	}

	// --- Copy path parameters ---
	if params := c.ctx.Params; len(params) > 0 {
		paramMap := make(map[string]any, len(params))
		for _, param := range params {
			if param.Key != "" && param.Value != "" {
				paramMap[param.Key] = param.Value
			}
		}
		data.PathParams = paramMap
	}
	// --- Copy cookies ---
	if cookies := c.ctx.Request.Cookies(); len(cookies) > 0 {
		cookieMap := make(map[string]any, len(cookies))
		for _, cookie := range cookies {
			if cookie.Name != "" && cookie.Value != "" {
				cookieMap[cookie.Name] = cookie.Value
			}
		}
		data.Cookies = cookieMap
	}

	return data
}

func (c *HttpContext) CommonLog(cmd, scenario string) {
	inComing := c.Incoming()

	xrid := inComing.Header["x-request-id"]
	var initInvoke string
	if xrid != nil {
		initInvoke = xrid.(string)
	}

	if initInvoke == "" {
		initInvoke = GenerateXTid("clnt")
	}

	detailLog, summaryLog := c.Log().NewLog(c.ctx.Request.Context(), initInvoke, scenario)

	protocol := c.ctx.Request.Proto
	protocolMethod := c.ctx.Request.Method
	detailLog.AddInputHttpRequest("client", cmd, initInvoke, inComing, true, protocol, protocolMethod)
	c.baseCommand = cmd
	c.initInvoke = initInvoke

	c.detailLog = detailLog
	c.summaryLog = summaryLog
}

func (c *HttpContext) DetailLog() logger.DetailLog {
	return c.detailLog
}
func (c *HttpContext) SummaryLog() logger.SummaryLog {
	return c.summaryLog
}

func (c *HttpContext) Context() context.Context {
	return c.ctx.Request.Context()
}

func (c *HttpContext) SendMessage(topic string, message any, opts ...OptionProducerMsg) (RecordMetadata, error) {
	ctx, span := otel.GetTracerProvider().Tracer("gokp").Start(c.Context(), "kafka-producer-"+topic)
	c.ctx.Request = c.ctx.Request.WithContext(ctx)
	defer span.End()
	invoke := uuid.NewString()
	c.detailLog.AddOutputRequest("kafka", "producer", invoke, message, map[string]any{
		"Body": map[string]any{
			"topic": topic,
			"value": message,
		},
	}, "kafka", "")
	result, err := producer(c.cfg.producer, topic, message, opts...)
	if err != nil {
		c.detailLog.AddInputResponse("kafka", "producer", invoke, message, err.Error())
		c.summaryLog.AddError("kafka", "producer", "", err.Error())
		return RecordMetadata{}, err
	}
	c.detailLog.AddInputResponse("kafka", "producer", invoke, result, result)
	c.summaryLog.AddSuccess("kafka", "producer", "20000", "success")
	return result, nil
}

func (c *HttpContext) Log() ILogger {
	switch logger := c.Context().Value(key).(type) {
	case ILogger:
		return logger
	default:
		return c.log
	}
}

func (c *HttpContext) Query(name string) string {
	return c.ctx.Query(name)
}

func (c *HttpContext) Param(name string) string {
	return c.ctx.Param(name)
}

func (c *HttpContext) ReadInput(data any) error {
	// Read the body into a byte slice
	err := json.NewDecoder(c.ctx.Request.Body).Decode(data)
	if err != nil {
		if c.copyBody != nil {
			if nErr := json.Unmarshal(c.copyBody, data); nErr != nil {
				return nErr
			}
			c.copyBody = nil
			return nil
		}
		return err
	}

	return nil
}

func (c *HttpContext) Response(responseCode int, responseData any) error {
	// c.w.Header().Set("Content-type", "application/json; charset=UTF8")
	// c.w.WriteHeader(responseCode)
	// return json.NewEncoder(c.w).Encode(responseData)
	c.ctx.JSON(responseCode, responseData)
	c.detailLog.AddOutputResponse("client", c.baseCommand, c.initInvoke, responseData, responseData)

	if !c.summaryLog.IsEnd() {
		c.summaryLog.End(fmt.Sprintf("%d", responseCode), "")
	}
	c = nil
	return nil

}

func (c *HttpContext) SetHeader(key, value string) {
	c.ctx.Header(key, value)
}

func (c *HttpContext) GetHeader(key string) string {
	return c.ctx.GetHeader(key)
}

func (c *HttpContext) Next() {
	c.ctx.Next()
}

func (c *HttpContext) Header() http.Header {
	return c.ctx.Request.Header
}

func (c *HttpContext) FullPath() string {
	return c.ctx.FullPath()
}
func (c *HttpContext) GetMethod() string {
	return c.ctx.Request.Method
}
func (c *HttpContext) GetPath() string {
	return c.ctx.Request.URL.Path
}
