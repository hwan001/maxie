package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// healthzHandler is a liveness probe: the process is up and serving. It does not
// check dependencies so it stays green during transient DB blips.
func healthzHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// readyzHandler is a readiness probe: reports 503 unless the file DB is
// reachable, so orchestrators can withhold traffic until dependencies are up.
func readyzHandler(c *gin.Context) {
	if fileDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "reason": "db_uninitialized"})
		return
	}
	if err := fileDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "reason": "db_unreachable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// requestLogger emits one structured (JSON) log line per request via slog,
// replacing Gin's default text logger. Health probes are dropped to keep logs
// signal-dense.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		if path == "/healthz" || path == "/readyz" {
			return
		}

		attrs := []any{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		msg := "request"
		switch {
		case c.Writer.Status() >= 500:
			slog.Error(msg, attrs...)
		case c.Writer.Status() >= 400:
			slog.Warn(msg, attrs...)
		default:
			slog.Info(msg, attrs...)
		}
	}
}
