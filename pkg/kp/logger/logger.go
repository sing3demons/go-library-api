package logger

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	ProjectName string
	Namespace   string
	AppLog      AppLog           `json:"appLog"`
	Summary     SummaryLogConfig `json:"summary"`
	Detail      DetailLogConfig  `json:"detail"`
}

type AppLog struct {
	Name       string        `json:"name"`
	LogFile    bool          `json:"logFile"`
	LogConsole bool          `json:"logConsole"`
	LogLevel   zapcore.Level `json:"logLevel"`
	AppLog     *zap.Logger
}

type SummaryLogConfig struct {
	Name       string `json:"name"`
	RawData    bool   `json:"rawData"`
	LogFile    bool   `json:"logFile"`
	LogConsole bool   `json:"logConsole"`
	LogSummary *zap.Logger
}

type DetailLogConfig struct {
	Name       string `json:"name"`
	RawData    bool   `json:"rawData"`
	LogFile    bool   `json:"logFile"`
	LogConsole bool   `json:"logConsole"`
	LogDetail  *zap.Logger
}

type InputOutputLog struct {
	Invoke   string      `json:"Invoke"`
	Event    string      `json:"Event"`
	Protocol *string     `json:"Protocol,omitempty"`
	Type     string      `json:"Type"`
	RawData  any `json:"RawData,omitempty"`
	Data     any `json:"Data"`
	ResTime  *string     `json:"ResTime,omitempty"`
}

type detailLog struct {
	LogType         string               `json:"LogType"`
	Host            string               `json:"Host"`
	AppName         string               `json:"AppName"`
	Instance        *string              `json:"Instance,omitempty"`
	Session         string               `json:"Session"`
	InitInvoke      string               `json:"InitInvoke"`
	Scenario        string               `json:"Scenario"`
	Identity        string               `json:"Identity"`
	InputTimeStamp  *string              `json:"InputTimeStamp,omitempty"`
	Input           []InputOutputLog     `json:"Input"`
	OutputTimeStamp *string              `json:"OutputTimeStamp,omitempty"`
	Output          []InputOutputLog     `json:"Output"`
	ProcessingTime  *string              `json:"ProcessingTime,omitempty"`
	conf            DetailLogConfig      `json:"-"`
	startTimeDate   time.Time            `json:"-"`
	inputTime       *time.Time           `json:"-"`
	outputTime      *time.Time           `json:"-"`
	timeCounter     map[string]time.Time `json:"-"`
	// req             *http.Request
	mu sync.Mutex
	autoEnd bool
}

type logEvent struct {
	node           string
	cmd            string
	invoke         string
	logType        string
	rawData        any
	data           any
	resTime        string
	protocol       string
	protocolMethod string
}

type summaryLog struct {
	mu            sync.Mutex
	requestTime   *time.Time
	session       string
	initInvoke    string
	cmd           string
	blockDetail   []BlockDetail
	optionalField OptionalFields
	conf          LogConfig
}

type SummaryResult struct {
	ResultCode string `json:"ResultCode"`
	ResultDesc string `json:"ResultDesc"`
	Count      int    `json:"-"`
}

type BlockDetail struct {
	Node   string          `json:"Node"`
	Cmd    string          `json:"Cmd"`
	Result []SummaryResult `json:"Result"`
	Count  int             `json:"Count"`
}

type OptionalFields map[string]any

type ContextKey string

const (
	TraceIDKey      ContextKey = "trace_id"
	SpanIDKey       ContextKey = "span_id"
	xSession        ContextKey = "session"
	ContentType                = "Content-Type"
	ContentTypeJSON            = "application/json"
	key                        = "logger"
	Summary                    = "Summary"
	Detail                     = "Detail"
)

func getModuleNameFromGoMod() string {
	file, err := os.Open("go.mod")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module") {
			moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			return filepath.Base(moduleName)
		}
	}

	if err := scanner.Err(); err != nil {
		return ""
	}

	return ""
}

var configLog LogConfig = LogConfig{
	ProjectName: getModuleNameFromGoMod(),
	Namespace:   "",
	AppLog: AppLog{
		Name:       "./logs/app",
		LogFile:    false,
		LogConsole: true,
	},
	Detail: DetailLogConfig{
		Name:       "./logs/detail",
		RawData:    true,
		LogFile:    false,
		LogConsole: false,
	},
	Summary: SummaryLogConfig{
		Name:       "./logs/summary",
		RawData:    true,
		LogFile:    false,
		LogConsole: false,
	},
}

func LoadLogConfig(cfg LogConfig) *LogConfig {
	if cfg.Namespace != "" {
		configLog.Namespace = cfg.Namespace
	}

	if cfg.ProjectName != "" {
		configLog.ProjectName = cfg.ProjectName
	}

	if cfg.AppLog.Name != "" {
		configLog.AppLog.Name = cfg.AppLog.Name
	}

	if cfg.AppLog.LogFile {
		configLog.AppLog.LogFile = cfg.AppLog.LogFile
		if err := ensureLogDirExists(configLog.AppLog.Name); err != nil {
			log.Fatal(err)
		}

		configLog.AppLog.AppLog = newLogFile(configLog.AppLog.Name)
	}

	if cfg.AppLog.LogConsole {
		configLog.AppLog.LogConsole = cfg.AppLog.LogConsole
	}

	if cfg.Detail.Name != "" {
		configLog.Detail.Name = cfg.Detail.Name
	}

	if cfg.Detail.RawData {
		configLog.Detail.RawData = cfg.Detail.RawData
	}

	if cfg.Detail.LogFile {
		configLog.Detail.LogFile = cfg.Detail.LogFile

		if err := ensureLogDirExists(configLog.Summary.Name); err != nil {
			log.Fatal(err)
		}

		configLog.Detail.LogDetail = newLogFile(configLog.Detail.Name)
	}

	if cfg.Detail.LogConsole {
		configLog.Detail.LogConsole = cfg.Detail.LogConsole
	}

	if cfg.Summary.Name != "" {
		configLog.Summary.Name = cfg.Summary.Name
	}

	if cfg.Summary.RawData {
		configLog.Summary.RawData = cfg.Summary.RawData
	}

	if cfg.Summary.LogConsole {
		configLog.Summary.LogConsole = cfg.Summary.LogConsole
	}

	if cfg.Summary.LogFile {
		configLog.Summary.LogFile = cfg.Summary.LogFile
		if err := ensureLogDirExists(configLog.Summary.Name); err != nil {
			log.Fatal(err)
		}

		configLog.Summary.LogSummary = newLogFile(configLog.Summary.Name)
	}

	return &configLog
}

func newLogFile(path string) *zap.Logger {
	log, err := createLogger(path)
	if err != nil {
		fmt.Println("Failed to create log file logger:", err)
	}
	return log
}

func ensureLogDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if path != "" {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create log directory[%s]: %v", path, err)
			}
		}
	}
	return nil
}

func createLogger(path string) (*zap.Logger, error) {
	// Create log file with rotating mechanism
	logFile := filepath.Join(path, getLogFileName(time.Now()))

	// Create a zapcore encoder config
	encCfg := zapcore.EncoderConfig{
		MessageKey:   "msg",
		TimeKey:      "time",
		LevelKey:     "level",
		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	// File encoder using console format
	fileEncoder := zapcore.NewConsoleEncoder(encCfg)

	// Setting up lumberjack logger for log rotation
	writerSync := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    500, // megabytes
		MaxBackups: 3,   // number of backups
		MaxAge:     1,   // days
		LocalTime:  true,
		Compress:   true, // compress the backups
	})

	// Create the core with InfoLevel logging
	core := zapcore.NewCore(fileEncoder, writerSync, zap.InfoLevel)

	// Create logger
	log := zap.New(core)

	return log, nil
}

func getLogFileName(t time.Time) string {
	appName := configLog.ProjectName
	year, month, day := t.Date()
	hour, minute, second := t.Clock()

	return fmt.Sprintf("%s_%04d%02d%02d_%02d%02d%02d.log", appName, year, month, day, hour, minute, second)
}