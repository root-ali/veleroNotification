package kubernetes

import (
	"go.uber.org/zap"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubernetesClientRepository interface {
	SendMessage(message, status string) error
}

type KubernetesClientService interface {
	HealthCheck() error
	VeleroBackupWatch()
}

type KubernetesClient struct {
	config        *rest.Config
	clientset     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	stopChan      chan struct{}
	kr            KubernetesClientRepository
	logger        *zap.SugaredLogger
}
