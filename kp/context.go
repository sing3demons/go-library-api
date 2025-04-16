package kp

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sing3demons/go-library-api/kp/logger"
)

type IContext interface {
	Context() context.Context
	SetHeader(key, value string)
	GetHeader(key string) string

	Log() ILogger
	Param(name string) string
	Query(name string) string
	ReadInput(data any) error
	Response(code int, data any) error

	// SendMessage(topic string, payload any, opts ...OptionProducerMsg) (RecordMetadata, error)
	CommonLog(cmd, initInvoke, scenario string)
	DetailLog() logger.DetailLog
	SummaryLog() logger.SummaryLog
}

type HandleFunc func(ctx IContext) error

type ServiceHandleFunc HandleFunc

type Middleware func(HandleFunc) HandleFunc

type HttpContext struct {
	ctx        *gin.Context
	cfg        *KafkaConfig
	log        ILogger
	detailLog  logger.DetailLog
	summaryLog logger.SummaryLog
	baseCommand string
	initInvoke string
}

func newMuxContext(c *gin.Context, cfg *KafkaConfig, log ILogger) IContext {
	ctx := InitSession(c.Request.Context(), log)
	c.Request = c.Request.WithContext(ctx)
	return &HttpContext{ctx: c, cfg: cfg, log: log}
}

func (c *HttpContext) CommonLog(cmd, initInvoke, scenario string) {
	detailLog, summaryLog := c.Log().NewLog(c.ctx.Request.Context(), initInvoke, scenario)
	// c.ctx.Request = c.ctx.Request.WithContext(context.WithValue(c.ctx.Request.Context(), key, c.log))

	detailLog.AddInputHttpRequest("client", cmd, initInvoke, c.ctx.Request.Clone(c.ctx.Request.Context()), true)
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
	return producer(c.cfg.producer, topic, message, opts...)
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
	// return json.NewDecoder(c.r.Body).Decode(data)
	return c.ctx.BindJSON(data)
}

func (c *HttpContext) Response(responseCode int, responseData any) error {
	// c.w.Header().Set("Content-type", "application/json; charset=UTF8")

	// c.w.WriteHeader(responseCode)

	// return json.NewEncoder(c.w).Encode(responseData)
	c.ctx.JSON(responseCode, responseData)
	c.detailLog.AddOutputRequest("client", c.baseCommand, c.initInvoke, responseData, responseData)
	c = nil
	return nil

}

func (c *HttpContext) SetHeader(key, value string) {
	c.ctx.Header(key, value)
}

func (c *HttpContext) GetHeader(key string) string {
	return c.ctx.GetHeader(key)
}
