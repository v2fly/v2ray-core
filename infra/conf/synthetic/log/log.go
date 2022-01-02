package log

import (
	"strings"

	"github.com/v2fly/v2ray-core/v5/app/log"
	clog "github.com/v2fly/v2ray-core/v5/common/log"
)

func DefaultLogConfig() *log.Config {
	return &log.Config{
		Access: &log.LogSpecification{Type: log.LogType_None},
		Error:  &log.LogSpecification{Type: log.LogType_Console, Level: clog.Severity_Warning},
	}
}

type LogConfig struct { // nolint: revive
	AccessLog string `json:"access"`
	ErrorLog  string `json:"error"`
	LogLevel  string `json:"loglevel"`
}

func (v *LogConfig) Build() *log.Config {
	if v == nil {
		return nil
	}
	config := &log.Config{
		Access: &log.LogSpecification{Type: log.LogType_Console},
		Error:  &log.LogSpecification{Type: log.LogType_Console},
	}

	if v.AccessLog == "none" {
		config.Access.Type = log.LogType_None
	} else if len(v.AccessLog) > 0 {
		config.Access.Path = v.AccessLog
		config.Access.Type = log.LogType_File
	}
	if v.ErrorLog == "none" {
		config.Error.Type = log.LogType_None
	} else if len(v.ErrorLog) > 0 {
		config.Error.Path = v.ErrorLog
		config.Error.Type = log.LogType_File
	}

	level := strings.ToLower(v.LogLevel)
	switch level {
	case "debug":
		config.Error.Level = clog.Severity_Debug
	case "info":
		config.Error.Level = clog.Severity_Info
	case "error":
		config.Error.Level = clog.Severity_Error
	case "none":
		config.Error.Type = log.LogType_None
		config.Error.Type = log.LogType_None
	default:
		config.Error.Level = clog.Severity_Warning
	}
	return config
}
