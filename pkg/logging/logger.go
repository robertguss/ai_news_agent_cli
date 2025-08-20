package logging

import (
        "errors"
        "fmt"
        "log"
        "os"
        "path/filepath"
        "time"

        "github.com/robertguss/ai-news-agent-cli/pkg/errs"
)

var logger *log.Logger

func Init(logFile string) error {
        if logFile == "" {
                logger = log.New(os.Stderr, "", log.LstdFlags)
                return nil
        }

        dir := filepath.Dir(logFile)
        if err := os.MkdirAll(dir, 0755); err != nil {
                return fmt.Errorf("create log directory: %w", err)
        }

        file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
        if err != nil {
                return fmt.Errorf("open log file: %w", err)
        }

        logger = log.New(file, "", log.LstdFlags)
        return nil
}

func Error(operation string, err error) {
        if logger == nil {
                return
        }

        timestamp := time.Now().Format("2006-01-02 15:04:05")
        
        var appErr *errs.AppError
        if errors.As(err, &appErr) {
                logger.Printf("[ERROR] %s - %s: %s (Type: %d, Retryable: %t)", 
                        timestamp, operation, appErr.Err.Error(), appErr.Type, appErr.Retryable)
        } else {
                logger.Printf("[ERROR] %s - %s: %s", timestamp, operation, err.Error())
        }
}

func Info(operation string, message string) {
        if logger == nil {
                return
        }

        timestamp := time.Now().Format("2006-01-02 15:04:05")
        logger.Printf("[INFO] %s - %s: %s", timestamp, operation, message)
}

func Warn(operation string, message string) {
        if logger == nil {
                return
        }

        timestamp := time.Now().Format("2006-01-02 15:04:05")
        logger.Printf("[WARN] %s - %s: %s", timestamp, operation, message)
}

func Retry(operation string, attempt int, err error) {
        if logger == nil {
                return
        }

        timestamp := time.Now().Format("2006-01-02 15:04:05")
        logger.Printf("[RETRY] %s - %s: Attempt %d failed: %s", 
                timestamp, operation, attempt, err.Error())
}
