package log // import "v2ray.com/core/common/log"

import (
	"fmt"
	"sync"

	"v2ray.com/core/common/serial"
)

// Message is the interface for all log messages.
type Message interface {
	String() string
}

// Handler is the interface for log handler.
type Handler interface {
	Handle(msg Message)
}

// GeneralMessage is a general log message that can contain all kind of content.
type GeneralMessage struct {
	Severity Severity
	Content  interface{}
}

// String implements Message.
func (m *GeneralMessage) String() string {
	return serial.Concat("[", m.Severity, "] ", m.Content)
}

// Record writes a message into log stream.
func Record(msg Message) {
	logHandler.Handle(msg)
}
func Info(format string, v ...interface{}){
	Record(&GeneralMessage{
		Severity: Severity_Info,
		Content: fmt.Sprintf(format, v...),
	})
}
func Debug(format string, v ...interface{}){
	Record(&GeneralMessage{
		Severity: Severity_Debug,
		Content: fmt.Sprintf(format, v...),
	})
}
func Warn(format string, v ...interface{}){
	Record(&GeneralMessage{
		Severity: Severity_Warning,
		Content: fmt.Sprintf(format, v...),
	})
}
func Error(format string, v ...interface{}){
	Record(&GeneralMessage{
		Severity: Severity_Error,
		Content: fmt.Sprintf(format, v...),
	})
}

var (
	logHandler syncHandler
)

// RegisterHandler register a new handler as current log handler. Previous registered handler will be discarded.
func RegisterHandler(handler Handler) {
	if handler == nil {
		panic("Log handler is nil")
	}
	logHandler.Set(handler)
}

type syncHandler struct {
	sync.RWMutex
	Handler
}

func (h *syncHandler) Handle(msg Message) {
	h.RLock()
	defer h.RUnlock()

	if h.Handler != nil {
		h.Handler.Handle(msg)
	}
}

func (h *syncHandler) Set(handler Handler) {
	h.Lock()
	defer h.Unlock()

	h.Handler = handler
}
