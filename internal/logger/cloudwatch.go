package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type CloudWatchLogger struct {
	logGroupName  string
	logStreamName string
}

func NewCloudWatchLogger(ctx context.Context, logGroupName, logStreamName string) (*CloudWatchLogger, error) {
	// Ensure we have a base stream name
	if strings.TrimSpace(logStreamName) == "" {
		logStreamName = "api-server"
	}

	// Attempt to enrich stream name with EC2 instance-id when available
	instanceID := ""
	// Use a short timeout for metadata lookup so local dev won't hang
	mdCtx, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(mdCtx, http.MethodGet, "http://169.254.169.254/latest/meta-data/instance-id", nil)
	if err == nil {
		client := &http.Client{Timeout: 750 * time.Millisecond}
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				if body, err := io.ReadAll(resp.Body); err == nil {
					instanceID = strings.TrimSpace(string(body))
				}
			}
		}
	}

	// If we couldn't get instance-id, fall back to hostname (useful for local dev)
	if instanceID == "" {
		if hn, err := os.Hostname(); err == nil {
			// shorten hostname if it contains dots
			if idx := strings.IndexByte(hn, '.'); idx > 0 {
				hn = hn[:idx]
			}
			instanceID = hn
		}
	}

	// Compose final stream name, avoid duplicating instance id if already present
	finalStream := logStreamName
	if instanceID != "" && !strings.Contains(logStreamName, instanceID) {
		finalStream = fmt.Sprintf("%s-%s", logStreamName, instanceID)
	}

	log.Printf("CloudWatch logger initialized for group: %s, stream: %s", logGroupName, finalStream)
	return &CloudWatchLogger{
		logGroupName:  logGroupName,
		logStreamName: finalStream,
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
