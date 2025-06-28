package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type FileLogger struct {
	logChannel chan string
	file       *os.File
}

func NewFileLogger(logFilepath string) (*FileLogger, error) {
	dir := filepath.Dir(logFilepath)

	// #nosec G301
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(
		logFilepath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644) // #nosec G302 G304
	if err != nil {
		return nil, err
	}

	logger := &FileLogger{
		logChannel: make(chan string, 1000),
		file:       file,
	}

	go logger.processLogs()
	return logger, nil
}

func (l *FileLogger) LogResponse(provider string, resp *http.Response) {
	if resp == nil || resp.Body == nil {
		l.log(fmt.Sprintf("%s - Empty response", provider))
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		l.log(fmt.Sprintf("%s - Failed to read body: %v", provider, err))
		return
	}

	l.log(fmt.Sprintf("%s - %s", provider, string(bodyBytes)))

	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
}

func (l *FileLogger) log(message string) {
	select {
	case l.logChannel <- message:
	default:
		log.Println("Log channel full, dropping log:", message)
	}
}

func (l *FileLogger) processLogs() {
	for logEntry := range l.logChannel {
		_, err := l.file.WriteString(logEntry + "\n")
		if err != nil {
			log.Printf("Error writing to log file: %v\n", err)
		}
	}
}
