package grohl

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
)

type Error struct {
	Message    string
	reportable bool
	InnerError error
	data       Data
	stack      []byte
}

type HttpError struct {
	StatusCode int
	*Error
}

// NewError wraps an error with the error's message.
func NewError(err error) *Error {
	return NewErrorf(err, "")
}

// NewErrorf wraps an error with a formatted message.
func NewErrorf(err error, format string, a ...interface{}) *Error {
	var msg string
	if len(format) > 0 {
		msg = fmt.Sprintf(format, a...)
	} else if err != nil {
		msg = err.Error()
	}

	return &Error{
		Message:    msg,
		reportable: true,
		InnerError: err,
		data:       nil,
		stack:      Stack(),
	}
}

// NewHttpError wraps an error with an HTTP status code and the given error's
// message.
func NewHttpError(err error, status int) *HttpError {
	return NewHttpErrorf(err, status, "")
}

// NewHttpErrorf wraps an error with an HTTP status code and a formatted message.
func NewHttpErrorf(err error, status int, format string, a ...interface{}) *HttpError {
	if status < 1 {
		status = 500
	}
	return &HttpError{status, NewErrorf(err, format, a...)}
}

// Error returns the error message.  This will be the inner error's message,
// unless a formatted message is provided from Errorf().
func (e *Error) Error() string {
	return e.Message
}

// Stack returns the runtime stack stored with this Error.
func (e *Error) Stack() []byte {
	return e.stack
}

// Data returns the error's current grohl.Data context.
func (e *Error) Data() Data {
	return e.data
}

// Reportable returns whether this error should be sent to the grohl
// ErrorReporter.
func (e *Error) Reportable() bool {
	return e.reportable
}

// Add adds the key and value to this error's context.
func (e *Error) Add(key string, value interface{}) {
	if e.data == nil {
		e.data = Data{}
	}
	e.data[key] = value
}

// Delete removes the key from this error's context.
func (e *Error) Delete(key string) {
	if e.data != nil {
		delete(e.data, key)
	}
}

// Stack returns the current runtime stack (up to 1MB).
func Stack() []byte {
	stackBuf := make([]byte, 1024*1024)
	written := runtime.Stack(stackBuf, false)
	return stackBuf[:written]
}

type ErrorReporter interface {
	Report(err error, data Data) error
}

// Report writes the error to the ErrorReporter, or logs it if there is none.
func (c *Context) Report(err error, data Data) error {
	if rErr, ok := err.(reportableError); ok {
		if rErr.Reportable() == false {
			return nil
		}
	}

	dataMaps := make([]Data, 1, 3)
	dataMaps[0] = c.Data()
	if gErr, ok := err.(grohlError); ok {
		if errData := gErr.Data(); errData != nil {
			dataMaps = append(dataMaps, errData)
		}
	}

	if data != nil {
		dataMaps = append(dataMaps, data)
	}

	merged := dupeMaps(dataMaps...)
	errorToMap(err, merged)

	if c.ErrorReporter != nil {
		return c.ErrorReporter.Report(err, merged)
	} else {
		var logErr error
		logErr = c.log(merged)
		if logErr != nil {
			return logErr
		}

		for _, line := range ErrorBacktraceLines(err) {
			lineData := dupeMaps(merged)
			lineData["site"] = line
			logErr = c.log(lineData)
			if logErr != nil {
				return logErr
			}
		}
		return nil
	}
}

// ErrorBacktrace creates a backtrace of the call stack.
func ErrorBacktrace(err error) string {
	return string(errorStack(err))
}

// ErrorBacktraceLines creates a backtrace of the call stack, split into lines.
func ErrorBacktraceLines(err error) []string {
	byteLines := bytes.Split(errorStack(err), byteLineBreak)
	lines := make([]string, 0, len(byteLines))

	// skip top two frames which are this method and `errorBacktraceBytes`
	for i := 2; i < len(byteLines); i++ {
		lines = append(lines, string(byteLines[i]))
	}
	return lines
}

type stackedError interface {
	Stack() []byte
}

type reportableError interface {
	Reportable() bool
}

type grohlError interface {
	Data() Data
}

func errorStack(err error) []byte {
	if sErr, ok := err.(stackedError); ok {
		return sErr.Stack()
	}
	return Stack()
}

func errorToMap(err error, data Data) {
	data["at"] = "exception"
	data["class"] = reflect.TypeOf(err).String()
	data["message"] = err.Error()
}

var byteLineBreak = []byte{'\n'}
