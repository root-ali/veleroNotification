package kubernetes

import (
	"context"
	"encoding/json"
	"strconv"

	vr_errors "github.com/root-ali/velero-reporter/pkg/errors"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func (kc *KubernetesClient) VeleroBackupWatch() {

	backupRes := schema.GroupVersionResource{
		Group:    "velero.io",
		Version:  "v1",
		Resource: "backups",
	}

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return kc.dynamicClient.Resource(backupRes).Namespace("velero").List(context.Background(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return kc.dynamicClient.Resource(backupRes).Namespace("velero").Watch(context.Background(), options)
		},
	}

	informer := cache.NewSharedInformer(
		lw,
		&unstructured.Unstructured{},
		0,
	)

	// Set up event handlers.
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{

		UpdateFunc: func(oldObj, newObj interface{}) {
			u := newObj.(*unstructured.Unstructured)
			kc.logger.Info("a backup is updated")
			err := kc.checkBackupUpdateStatus(u)
			if err != nil {
				kc.logger.Errorw("Cannot Handle backup properly")
			}
		},
	})

	go informer.Run(kc.stopChan)

}

func (kc *KubernetesClient) Stop() {
	close(kc.stopChan)
}

func (kc *KubernetesClient) checkBackupUpdateStatus(u *unstructured.Unstructured) error {
	resourceVersion, err := strconv.Atoi(u.GetResourceVersion())
	if err != nil {
		kc.logger.Errorw("Cannot convert resource version to int")
		return vr_errors.VELERO_CANNOT_CONVERT_RESOURCE_VERSION_TO_INT
	}
	oldResourceVersion, err := kc.getLastResourceVesrion()
	if err != nil {
		kc.logger.Errorw("Cannot get latest resource version")
		return err
	}
	if oldResourceVersion < resourceVersion {

		status, err := kc.getBackupStatus(u)
		if err != nil {
			kc.logger.Error()
		}
		err = kc.checkBackupStatus(status, u.GetNamespace(), u.GetName())
		if err != nil {
			kc.logger.Error("error checking backup status", err)
		} else {
			kc.logger.Info("backup is updated")
			err = kc.updateConfigMap(strconv.Itoa(resourceVersion))
			if err != nil {
				kc.logger.Errorw("Cannot update configmap ", "error", err)
				return vr_errors.VELERO_UPDATE_CONFIGMAP_ERROR
			}
		}
		return nil
	} else {
		kc.logger.Info("We are checking an older version backup nothing to do")
		return nil
	}

}

func (kc *KubernetesClient) updateConfigMap(resourceVersion string) error {
	configMap, err := kc.clientset.CoreV1().ConfigMaps("velero").Get(context.TODO(),
		"velero-notification-backup-last-resource-version", metav1.GetOptions{})
	if err != nil {
		kc.logger.Errorw("Cannot get configmap", "error", err)
		return vr_errors.VELERO_ERROR_RETIEVIE_CONFIGMAP
	}
	configMap.Data["resourceVersion"] = resourceVersion
	_, err = kc.clientset.CoreV1().ConfigMaps("velero").Update(context.TODO(),
		configMap, metav1.UpdateOptions{})
	if err != nil {
		kc.logger.Errorw("Cannot update configmap", "error", err)
		return vr_errors.VELERO_UPDATE_CONFIGMAP_ERROR
	}
	return nil
}

func (kc *KubernetesClient) getLastResourceVesrion() (int, error) {
	configMap, err := kc.clientset.CoreV1().ConfigMaps("velero").Get(context.TODO(),
		"velero-notification-backup-last-resource-version", metav1.GetOptions{})
	if err != nil {
		kc.logger.Error("Cannot get config map")
		return 0, vr_errors.VELERO_ERROR_RETIEVIE_CONFIGMAP
	}
	resourceVersion := configMap.GetResourceVersion()
	if resourceVersion == "" {
		kc.logger.Errorw("Resource version is null")
		return 0, vr_errors.VELERO_RESOURCEVERSION_IS_NULL
	}
	resourceVersionInt, err := strconv.Atoi(resourceVersion)
	if err != nil {
		kc.logger.Errorw("Cannot convert resource version to int")
		return 0, vr_errors.VELERO_CANNOT_CONVERT_RESOURCE_VERSION_TO_INT
	}
	return resourceVersionInt, nil
}

func (kc *KubernetesClient) getBackupStatus(u *unstructured.Unstructured) (v1.BackupStatus, error) {
	var status v1.BackupStatus

	statusRaw, ok, err := unstructured.NestedFieldNoCopy(u.Object, "status")
	if err != nil {
		kc.logger.Error("error retrieving status from velero", "error", err)
		return status, vr_errors.VELERO_RETIERIVE_STATUS_ERROR
	}
	if !ok {
		kc.logger.Errorw("status is missing", "error", vr_errors.VELERO_STATUS_MISSING)
		return status, vr_errors.VELERO_STATUS_MISSING
	}
	statusJSON, err := json.Marshal(statusRaw)
	if err != nil {
		kc.logger.Errorw("error marshalling status", "error", err)
		return status, vr_errors.VELERO_CANNOT_MARSHALL_STATUS
	}
	if err := json.Unmarshal(statusJSON, &status); err != nil {
		kc.logger.Errorw("Cannot unmarshall status", "error", err)
		return status, vr_errors.VELERO_CANNOT_MARSHALL_STATUS
	}
	return status, nil
}

func (kc *KubernetesClient) checkBackupStatus(status v1.BackupStatus, namespace string, name string) error {
	if status.Phase == "InProgress" {
		kc.logger.Infow("Backup is in InProgress mode nothing to do")
		return vr_errors.VELERO_BACKUP_NOT_COMPLETED
	} else if status.Phase == "Failed" {
		kc.logger.Infow("Backup Failed ", "status", status.Phase,
			"Failed Reason ", status.FailureReason)
		messsage := "Backup Failed: " + name + " in namespace " + namespace +
			"\n for reason: " + status.FailureReason +
			"\n Please run `velero backup logs " + name + "` for more information"
		err := kc.kr.SendMessage(messsage, "Failed")
		if err != nil {
			kc.logger.Errorw("Cannot send message ", "error", err)
			return err
		}
		return vr_errors.VELERO_BACKUP_NOT_COMPLETED
	} else if status.Phase == "Completed" {
		kc.logger.Infow("Backup Completed", "status", status.Phase,
			"Items", status.Progress.ItemsBackedUp, "Time", status.CompletionTimestamp)
		msg := "Backup Completed " + name + " on namespace " + namespace +
			"\n Please run `velero backup logs " + name + "` for more information"
		err := kc.kr.SendMessage(msg, "Success")
		if err != nil {
			kc.logger.Errorw("Cannot send message ", "error", err)
			return err
		}
		return nil
	} else {
		kc.logger.Infow("Backup is in different mode", "status", status.Phase)
		return vr_errors.VELERO_BACKUP_NOT_COMPLETED
	}
}
