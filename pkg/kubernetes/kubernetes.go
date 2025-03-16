package kubernetes

import (
	"context"
	vr_errors "github.com/root-ali/velero-reporter/pkg/errors"
	"go.uber.org/zap"
	config_v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	err = kc.initiateConfigMap()
	if err != nil {
		l.Errorw("Error initiating configmap", "error", err)
		panic(err)
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

func (kc *KubernetesClient) initiateConfigMap() error {
	configMap := &config_v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "velero-notification-backup-last-resource-version",
		},
		Data: map[string]string{
			"resourceVersion": "000000",
		},
	}
	create, err := kc.clientset.CoreV1().ConfigMaps("velero").Create(context.TODO(), configMap, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			kc.logger.Info("Configmap already exists")
			return nil
		}
		kc.logger.Errorw("Error creating configmap", "error", err)
		return vr_errors.KUBERNETES_CREATE_CONFIGMAP_ERROR
	}
	kc.logger.Infow("Initiate configmap for velero", "configmap.Name", create.Name, "configmap.Namespace", create.Namespace)
	return nil
}
