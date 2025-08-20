package errs

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"syscall"
)

// ErrorType represents different categories of errors for classification and handling.
type ErrorType int

const (
	TypeUnknown ErrorType = iota
	TypeNetwork
	TypeDatabase
	TypeAI
	TypeValidation
	TypeTimeout
	TypeRateLimit
)

type AppError struct {
	Type      ErrorType
	Operation string
	Err       error
	Retryable bool
}

func (e *AppError) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("%s: %v", e.Operation, e.Err)
	}
	return e.Err.Error()
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Wrap wraps an error with operation context and automatic error classification.
// It returns nil if the input error is nil, otherwise creates an AppError with context.
func Wrap(operation string, err error) error {
	if err == nil {
		return nil
	}

	appErr := &AppError{
		Operation: operation,
		Err:       err,
		Type:      classifyError(err),
	}

	appErr.Retryable = isRetryable(appErr.Type, err)
	return appErr
}

func classifyError(err error) ErrorType {
	if IsNetwork(err) {
		return TypeNetwork
	}
	if IsDatabase(err) {
		return TypeDatabase
	}
	if IsAI(err) {
		return TypeAI
	}
	if IsValidation(err) {
		return TypeValidation
	}
	if IsTimeout(err) {
		return TypeTimeout
	}
	if IsRateLimit(err) {
		return TypeRateLimit
	}
	return TypeUnknown
}

func isRetryable(errType ErrorType, err error) bool {
	switch errType {
	case TypeNetwork:
		return !IsInvalidURL(err)
	case TypeDatabase:
		return IsDBBusy(err)
	case TypeAI:
		return IsTransientAI(err)
	case TypeTimeout:
		return true
	case TypeRateLimit:
		return true
	default:
		return false
	}
}

func IsNetwork(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	var syscallErr *syscall.Errno
	if errors.As(err, &syscallErr) {
		switch *syscallErr {
		case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.ETIMEDOUT:
			return true
		}
	}

	errStr := strings.ToLower(err.Error())
	networkKeywords := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"no such host",
		"network unreachable",
		"timeout",
		"temporary failure",
	}

	for _, keyword := range networkKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}

	return false
}

func IsDatabase(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	dbKeywords := []string{
		"database",
		"sqlite",
		"sql",
		"constraint",
		"locked",
		"busy",
	}

	for _, keyword := range dbKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}

	return false
}

func IsDBBusy(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "database is locked") ||
		strings.Contains(errStr, "database is busy")
}

func IsAI(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	aiKeywords := []string{
		"gemini",
		"api key",
		"api_key",
		"quota",
		"rate limit",
		"generate content",
		"generative-ai",
	}

	for _, keyword := range aiKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}

	return false
}

func IsTransientAI(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	transientKeywords := []string{
		"rate limit",
		"quota exceeded",
		"server error",
		"service unavailable",
		"timeout",
		"temporary",
	}

	for _, keyword := range transientKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}

	return false
}

func IsValidation(err error) bool {
	if err == nil {
		return false
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	errStr := strings.ToLower(err.Error())
	validationKeywords := []string{
		"invalid url",
		"malformed",
		"parse error",
		"validation",
	}

	for _, keyword := range validationKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}

	return false
}

func IsInvalidURL(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "invalid url")
}

func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded")
}

func IsRateLimit(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many requests")
}

func IsRetryable(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Retryable
	}

	return IsNetwork(err) || IsDBBusy(err) || IsTransientAI(err) ||
		IsTimeout(err) || IsRateLimit(err)
}

func GetUserFriendlyMessage(err error) string {
	if err == nil {
		return ""
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Type {
		case TypeNetwork:
			if IsInvalidURL(err) {
				return "Invalid URL provided. Please check the URL format."
			}
			return "Network connection failed. Please check your internet connection and try again."
		case TypeDatabase:
			if IsDBBusy(err) {
				return "Database is temporarily busy. The operation will be retried automatically."
			}
			return "Database error occurred. Please check file permissions and disk space."
		case TypeAI:
			if IsRateLimit(err) {
				return "AI service rate limit reached. Please wait a moment and try again."
			}
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "api key") || strings.Contains(errStr, "api_key") {
				return "AI API key is missing or invalid. Please set GEMINI_API_KEY environment variable."
			}
			return "AI processing failed. The article will be saved without analysis."
		case TypeValidation:
			return "Invalid input provided. Please check your configuration."
		case TypeTimeout:
			return "Operation timed out. Please try again or check your network connection."
		case TypeRateLimit:
			return "Rate limit exceeded. Please wait a moment before trying again."
		default:
			return fmt.Sprintf("An error occurred: %v", appErr.Err)
		}
	}

	return fmt.Sprintf("An unexpected error occurred: %v", err)
}
