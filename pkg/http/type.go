package http

import (
	"github.com/root-ali/velero-reporter/pkg/health"
	"go.uber.org/zap"
)

type HttpHandler struct {
	hs health.HealthService
	l  *zap.SugaredLogger
}

func NewHttpService(hs health.HealthService, l *zap.SugaredLogger) *HttpHandler {
	return &HttpHandler{
		hs: hs,
		l:  l,
	}
}
