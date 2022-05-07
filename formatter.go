package logger

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	PrefixRequestIncoming = ">>>"
	PrefixRequestHandling = "==="
	PrefixRequestOutgoing = "<<<"

	RequestIdKey = "X-Request-ID"

	defaultTimeFormat    = time.RFC3339
	defaultDataSeparator = " |"

	logTimeKey  = "<time>"
	logLevelKey = "<level>"
	logIdKey    = "<id>"
	logMsgKey   = "<msg>"
)

var (
	defaultMsgFormat = fmt.Sprintf("[%v] [%v] [%v] %v", logTimeKey, logLevelKey, logIdKey, logMsgKey)

	FormatIdentifier = func(identifier string) string {
		return fmt.Sprintf("<%s>", identifier)
	}
)

type customFormatter struct {
	isDefault         bool
	jsonOutput        bool
	textDataSeparator string
	MsgPrefix         string
	MsgFormat         string
	TimeFormat        string
}

// DefaultFormatter uses default format defined in defaultLogFormat.
func DefaultFormatter(prefix string) *customFormatter {
	return &customFormatter{
		isDefault:         true,
		jsonOutput:        false,
		textDataSeparator: defaultDataSeparator,
		MsgPrefix:         prefix,
		MsgFormat:         defaultMsgFormat,
		TimeFormat:        defaultTimeFormat,
	}
}

// NewFormatter gives the flexibility to customized log format using
// time, level, data of logrus.Entry.
func NewFormatter() *customFormatter {
	return &customFormatter{
		TimeFormat: defaultTimeFormat,
	}
}

// SetLogFormat defined the message pattern for each log.
func (f *customFormatter) SetLogFormat(format string) *customFormatter {
	f.MsgFormat = format
	return f
}

// SetLogPrefix defined the message prefix for each log.
func (f *customFormatter) SetLogPrefix(prefix string) *customFormatter {
	f.MsgPrefix = prefix
	return f
}

// SetDataSeparator defined the separator to separate each key-value pair in logrus.Entry.Data
// for better readability in text-output.
func (f *customFormatter) SetDataSeparator(sep string) *customFormatter {
	f.textDataSeparator = sep
	return f
}

// SetJsonOutput enable whole log output into json format.
func (f *customFormatter) SetJsonOutput() *customFormatter {
	f.jsonOutput = true
	return f
}

// SetTimeFormat defined the output time format.
func (f *customFormatter) SetTimeFormat(format string) *customFormatter {
	f.TimeFormat = format
	return f
}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if f.MsgFormat == "" {
		return (&logrus.TextFormatter{}).Format(entry)
	}
	if f.jsonOutput {
		return f.jsonFormat(entry)
	}
	return f.textFormat(entry)
}

func (f *customFormatter) getMessage(entry *logrus.Entry) string {
	output := f.MsgFormat
	output = strings.Replace(output, logTimeKey, entry.Time.Format(f.TimeFormat), 1)
	output = strings.Replace(output, logLevelKey, fmt.Sprintf("%5s", strings.ToUpper(entry.Level.String())), 1)

	if id, exist := entry.Data[RequestIdKey]; exist {
		output = strings.Replace(output, logIdKey, id.(string), 1)
	} else if f.isDefault {
		output = strings.Replace(output, fmt.Sprintf(" [%v]", logIdKey), "", 1)
	}

	// To support other custom formats using data keys to design log output.
	for k, val := range entry.Data {
		if k == RequestIdKey {
			continue
		}

		if strings.Contains(output, FormatIdentifier(k)) {
			output = strings.Replace(output, FormatIdentifier(k), fmt.Sprintf("%v", val), 1)
			continue
		}
	}

	if entry.Message != "" {
		output = strings.Replace(output, logMsgKey, entry.Message, 1)
	} else {
		output = strings.Replace(output, fmt.Sprintf(" [%v]", logMsgKey), "", 1)
	}

	return output
}

// TODO: Customizable json structure.
type jsonOutputFormat struct {
	Timestamp string        `json:"timestamp"`
	Level     string        `json:"level"`
	ID        *string       `json:"id,omitempty"`
	Message   string        `json:"message"`
	Data      logrus.Fields `json:"data,omitempty"`
}

func (f *customFormatter) jsonFormat(entry *logrus.Entry) ([]byte, error) {
	format := jsonOutputFormat{
		Timestamp: entry.Time.Format(defaultTimeFormat),
		Level:     strings.ToUpper(entry.Level.String()),
		Message:   f.getMessage(entry),
		Data:      entry.Data,
	}
	if val, exist := entry.Data[RequestIdKey]; exist {
		id := val.(string)
		format.ID = &id
	}
	output, err := json.Marshal(&format)
	return []byte(fmt.Sprintln(output)), err
}

func (f *customFormatter) textFormat(entry *logrus.Entry) ([]byte, error) {
	output := f.getMessage(entry)
	for k, val := range entry.Data {
		if k == RequestIdKey {
			continue
		}

		if strings.Contains(output, FormatIdentifier(k)) {
			continue
		}

		// TODO: To redesign
		switch val.(type) {
		case map[string]interface{}:
			jsonVal, err := json.Marshal(val)
			if err == nil {
				output = fmt.Sprintf("%v%s %s = %v", output, f.textDataSeparator, k, jsonVal)
				continue
			}
		}

		output = fmt.Sprintf("%v%s %s = %v", output, f.textDataSeparator, k, val)
	}
	return []byte(fmt.Sprintln(output)), nil
}
