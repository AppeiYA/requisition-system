package config

import (
	"errors"
	"fmt"
	"strings"
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"

	LogFormatJSON = "json"
	LogFormatText = "text"
)

var (
	ErrLogLevelUnsupported  = errors.New("log level must be one of: debug, info, warn, error")
	ErrLogFormatUnsupported = errors.New("log format must be one of: json, text")
)

type LogConfig struct {
	Level     string
	Format    string
	AddSource bool
}

func LoadLogConfig(appEnv Environment) (*LogConfig, error) {
	level := strings.ToLower(getEnv("LOG_LEVEL", LogLevelInfo))
	switch level {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
	default:
		return nil, fmt.Errorf("%w, got %q", ErrLogLevelUnsupported, level)
	}

	defaultFormat := LogFormatText
	if appEnv == ProdEnv {
		defaultFormat = LogFormatJSON
	}
	format := strings.ToLower(getEnv("LOG_FORMAT", defaultFormat))
	switch format {
	case LogFormatJSON, LogFormatText:
	default:
		return nil, fmt.Errorf("%w, got %q", ErrLogFormatUnsupported, format)
	}

	addSource, err := getBoolEnv("LOG_ADD_SOURCE", false)
	if err != nil {
		return nil, err
	}

	return &LogConfig{
		Level:     level,
		Format:    format,
		AddSource: addSource,
	}, nil
}
