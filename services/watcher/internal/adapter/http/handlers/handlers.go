package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Health struct {
}

func NewHealth() *Health {
	return &Health{}
}

func (h *Health) AttachGinMux(mux *gin.Engine) {
	mux.Any("/system/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

type Metrics struct {
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (h *Metrics) AttachGinMux(mux *gin.Engine) {
	mux.Any("/system/metrics", func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})
}
