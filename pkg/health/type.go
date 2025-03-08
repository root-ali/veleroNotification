package health

import (
	"github.com/root-ali/velero-reporter/pkg/kubernetes"
	"go.uber.org/zap"
)

type HealthRepository interface {
	HealthCheck() error
}

type HealthService interface {
	Ready() error
	Healthy() error
}

type healthService struct {
	kc *kubernetes.KubernetesClient
	l  *zap.SugaredLogger
}
