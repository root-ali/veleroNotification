package kubernetes

import (
	"context"
	vr_errors "github.com/root-ali/velero-reporter/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
	"time"
)

func NewKubernetesClient(l *zap.SugaredLogger, kubeConfigPath string, kr KubernetesClientRepository) *KubernetesClient {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			l.Fatalf("Error building kubeconfig: %v", err)
		}
	}
	config.QPS = 100
	config.Burst = 200

	config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(config.QPS, config.Burst)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		l.Fatalf("Error creating clientset: %v", err.Error())
	}
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		l.Infow("Cannot create dynamic client", "error", err)
	}
	stopChan := make(chan struct{})
	kc := &KubernetesClient{
		config:        config,
		clientset:     clientset,
		dynamicClient: dynClient,
		stopChan:      stopChan,
		kr:            kr,
		logger:        l,
	}
	kc.VeleroBackupWatch()
	return kc
}

func (kc *KubernetesClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	kc.logger.Info("About to call /healthz kubernets api")
	result := kc.clientset.RESTClient().Get().AbsPath("/healthz").Do(ctx)
	if result.Error() != nil {
		kc.logger.Error("Error checking health:", "error", vr_errors.KUBERNETES_HEALTH_ERROR)
		return vr_errors.KUBERNETES_HEALTH_ERROR
	}

	var statusCode int
	result.StatusCode(&statusCode)
	kc.logger.Infow("Health check response is, ", "statusCode", statusCode)
	if statusCode != 200 {
		kc.logger.Errorw("Server not ready, status code: ", "statusCode", statusCode, "error", result.Error())
		return vr_errors.KUBERNETES_API_NOT_READY
	}

	body, err := result.Raw()
	if err != nil {
		kc.logger.Errorw("Error getting response body: %v", "error", err)
		return vr_errors.KUBERNETES_API_ERROR
	}
	if string(body) != "ok" {
		kc.logger.Errorw("Unexpected response body:", "error", err)
		return vr_errors.KUBERNETES_API_ERROR
	}

	return nil
}

	return nil
}
