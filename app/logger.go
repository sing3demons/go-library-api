package app

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/sing3demons/go-library-api/app/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ILogger interface {
	Sync() error
	Debug(args ...any)
	Debugf(template string, args ...any)
	Info(args ...any)
	Infof(template string, args ...any)
	Warn(args ...any)
	Warnf(template string, args ...any)
	WarnMsg(msg string, err error)
	Error(args ...any)
	Errorf(template string, args ...any)
	Err(msg string, err error)
	DPanic(args ...any)
	DPanicf(template string, args ...any)
	// Fatal(args ...any)
	// Fatalf(template string, args ...any)
	Printf(template string, args ...any)
	WithName(name string)
	Println(v ...any)

	Session(v string) ILogger

	NewLog(session, initInvoke, scenario string) (detailLog logger.DetailLog, summaryLog logger.SummaryLog)
}

const (
	TraceIDKey      ContextKey = "trace_id"
	SpanIDKey       ContextKey = "span_id"
	xSession        ContextKey = "session"
	ContentType                = "Content-Type"
	ContentTypeJSON            = "application/json"
	key             ContextKey = "logger"
	Summary                    = "Summary"
	Detail                     = "Detail"
)

func InitSession(c context.Context, l ILogger) context.Context {
	// get session from context
	session := c.Value(xSession)
	if session == nil {
		uuidV7, err := uuid.NewV7()
		if err != nil {
			uuidV7 = uuid.New()
		}
		session = uuidV7.String()

		// set session to context
		c = context.WithValue(c, xSession, session)
	}

	ctx := context.WithValue(c, key, l.Session(c.Value(xSession).(string)))

	return ctx
}

type Logger struct {
	log *zap.Logger
	ctx context.Context
}

type LogConfig struct {
	devMode  bool
	encoding string
}

func NewAppLogger(log ...*zap.Logger) ILogger {
	var logger *zap.Logger
	l := &LogConfig{}
	logWriter := zapcore.AddSync(os.Stdout)
	logLevel := zapcore.InfoLevel

	var encoderCfg zapcore.EncoderConfig
	if l.devMode {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
		logLevel = zapcore.DebugLevel
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	var encoder zapcore.Encoder
	encoderCfg.NameKey = "[SERVICE]"
	encoderCfg.TimeKey = "[TIME]"
	encoderCfg.LevelKey = "[LEVEL]"
	encoderCfg.FunctionKey = "[CALLER]"
	encoderCfg.CallerKey = "[LINE]"
	encoderCfg.MessageKey = "[MESSAGE]"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	encoderCfg.EncodeName = zapcore.FullNameEncoder
	encoderCfg.EncodeDuration = zapcore.StringDurationEncoder

	if l.encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	if len(log) > 0 {
		logger = log[0]
	} else {
		core := zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(logLevel))
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	}

	return &Logger{
		log: logger,
	}
}
func (l *Logger) L(c context.Context) ILogger {
	switch logger := c.Value(key).(type) {
	case ILogger:
		return logger
	default:
		return l
	}
}

func (l *Logger) Ctx() context.Context {
	return l.ctx
}

func (l *Logger) Sync() error {
	return l.log.Sync()
}

func (l Logger) Session(v string) ILogger {
	log := l.log.With(zap.String("session", v))
	l.log = log
	return &Logger{
		log: l.log,
	}
}

func (l *Logger) Debug(args ...any) {
	l.log.Sugar().Debug(args...)
}

func (l *Logger) Debugf(template string, args ...any) {
	l.log.Sugar().Debugf(template, args...)
}

func (l *Logger) Info(args ...any) {
	l.log.Sugar().Info(args...)
}

func (l *Logger) Infof(template string, args ...any) {
	l.log.Sugar().Infof(template, args...)
}

func (l *Logger) Warn(args ...any) {
	l.log.Sugar().Warn(args...)
}

func (l *Logger) Warnf(template string, args ...any) {
	l.log.Sugar().Warnf(template, args...)
}

func (l *Logger) WarnMsg(msg string, err error) {
	l.log.Warn(msg, zap.Error(err))
}

func (l *Logger) Error(args ...any) {
	l.log.Sugar().Error(args...)
}

func (l *Logger) Errorf(template string, args ...any) {
	l.log.Sugar().Errorf(template, args...)
}

func (l *Logger) Err(msg string, err error) {
	l.log.Error(msg, zap.Error(err))
}

func (l *Logger) DPanic(args ...any) {
	l.log.Sugar().DPanic(args...)
}

func (l *Logger) DPanicf(template string, args ...any) {
	l.log.Sugar().DPanicf(template, args...)
}

func (l *Logger) Fatal(args ...any) {
	l.log.Sugar().Fatal(args...)
}

func (l *Logger) Fatalf(template string, args ...any) {
	l.log.Sugar().Fatalf(template, args...)
}

func (l *Logger) Printf(template string, args ...any) {
	l.log.Sugar().Infof(template, args...)
}

func (l *Logger) WithName(name string) {
	l.log.Sugar().Named(name)
}

func (l *Logger) Println(v ...any) {
	l.log.Sugar().Info(v...)
}

func (l *Logger) NewLog(session, initInvoke, scenario string) (detailLog logger.DetailLog, summaryLog logger.SummaryLog) {
	detailLog = logger.NewDetailLog(session, initInvoke, scenario)
	summaryLog = logger.NewSummaryLog(session, initInvoke, scenario)
	return detailLog, summaryLog
}
