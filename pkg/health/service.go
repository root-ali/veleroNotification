package health

import (
	"github.com/root-ali/velero-reporter/pkg/kubernetes"
	"go.uber.org/zap"
)

func NewHealthService(kc *kubernetes.KubernetesClient, l *zap.SugaredLogger) HealthService {
	return &healthService{
		kc: kc,
		l:  l,
	}
}

func (h *healthService) Ready() error {
	return nil
}

func (h *healthService) Healthy() error {
	return h.kc.HealthCheck()
}
