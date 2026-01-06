package logger

import (
	"context"
	"fmt"
	"log"
	"time"
)

// CloudWatchLogger handles logging to CloudWatch Logs
// For now, this is a simple wrapper that logs to stdout
// In production, this would use the AWS CloudWatch Logs API
type CloudWatchLogger struct {
	logGroupName  string
	logStreamName string
}

// NewCloudWatchLogger creates a new CloudWatch logger
func NewCloudWatchLogger(ctx context.Context, logGroupName, logStreamName string) (*CloudWatchLogger, error) {
	log.Printf("CloudWatch logger initialized for group: %s, stream: %s", logGroupName, logStreamName)
	return &CloudWatchLogger{
		logGroupName:  logGroupName,
		logStreamName: logStreamName,
	}, nil
}

// LogRequest logs an HTTP request to CloudWatch
func (cwl *CloudWatchLogger) LogRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration, remoteAddr string) error {
	message := fmt.Sprintf(
		"[%s] %s %s | Status: %d | Duration: %dms | Remote: %s | Group: %s | Stream: %s",
		time.Now().Format("2006-01-02 15:04:05"),
		method,
		path,
		statusCode,
		duration.Milliseconds(),
		remoteAddr,
		cwl.logGroupName,
		cwl.logStreamName,
	)
	log.Printf("[CLOUDWATCH] %s", message)
	return nil
}

// LogError logs an error to CloudWatch
func (cwl *CloudWatchLogger) LogError(ctx context.Context, method, path string, err error, statusCode int) error {
	message := fmt.Sprintf(
		"[ERROR] [%s] %s %s | Status: %d | Error: %v | Group: %s | Stream: %s",
		time.Now().Format("2006-01-02 15:04:05"),
		method,
		path,
		statusCode,
		err,
		cwl.logGroupName,
		cwl.logStreamName,
	)
	log.Printf("[CLOUDWATCH] %s", message)
	return nil
}

// LogMessage logs a custom message to CloudWatch
func (cwl *CloudWatchLogger) LogMessage(ctx context.Context, message string) error {
	msg := fmt.Sprintf("[%s] %s | Group: %s | Stream: %s", time.Now().Format("2006-01-02 15:04:05"), message, cwl.logGroupName, cwl.logStreamName)
	log.Printf("[CLOUDWATCH] %s", msg)
	return nil
}
