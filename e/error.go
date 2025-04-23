package e

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type Error struct {
	IsPublic bool
	Code     string
	Message  string
	Cause    error
	File     string
	Line     int
}

func (e *Error) Error() string {
	if !e.IsPublic {
		return fmt.Sprintf("%s [%s]", e.Message, e.Code)
	}

	if e.Cause != nil {
		return fmt.Sprintf("%s [%s:%d]", e.Cause.Error(), e.File, e.Line)
	}

	return fmt.Sprintf("%s [%s:%d]", e.Message, e.File, e.Line)
}

func (e *Error) MarshalText() ([]byte, error) {
	return []byte(e.Error()), nil
}

func newError(message string, isPublic bool, cause error) error {
	err := &Error{
		Message:  message,
		Cause:    cause,
		IsPublic: isPublic,
	}

	if c, ok := cause.(*Error); ok {
		err.Code = c.Code
		err.Cause = c.Cause
		err.File = c.File
		err.Line = c.Line
	} else {
		_, file, line, _ := runtime.Caller(1)
		err.File = file
		err.Line = line
		filename := strings.ReplaceAll(base32.HexEncoding.EncodeToString([]byte(filepath.Base(file))), "=", "W")
		hashPrefix := strings.ToUpper(hex.EncodeToString(sha1.New().Sum([]byte(file)))[0:6])
		err.Code = fmt.Sprintf("%s%s-%d", hashPrefix, filename, line)
	}

	return err
}

func New(message string) error {
	return newError(message, false, nil)
}

func NewPublic(message string) error {
	return newError(message, true, nil)
}

func NewError(message string, cause error) error {
	return newError(message, false, cause)
}

func NewPublicError(message string, cause error) error {
	return newError(message, true, cause)
}
