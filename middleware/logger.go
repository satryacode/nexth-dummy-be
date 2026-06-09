// middleware/logger.go
package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type responseWriter struct {
	gin.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read body (intentionally logs full body — sensitive data exposure vuln)
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		elapsed := time.Since(start).Milliseconds()

		// Compact body for logging
		compactBody := string(bodyBytes)
		var compacted bytes.Buffer
		if json.Compact(&compacted, bodyBytes) == nil {
			compactBody = compacted.String()
		}

		logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.String("body", compactBody),
			zap.Int64("response_time_ms", elapsed),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}

func NewLogger() (*zap.Logger, error) {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		return nil, fmt.Errorf("create logs dir: %w", err)
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "timestamp"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	enc := zapcore.NewJSONEncoder(encCfg)

	logFile, err := os.OpenFile(
		"logs/requests.jsonl",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	core := zapcore.NewTee(
		zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), zap.InfoLevel),
		zapcore.NewCore(enc, zapcore.AddSync(logFile), zap.InfoLevel),
	)
	return zap.New(core), nil
}
