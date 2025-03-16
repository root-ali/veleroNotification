package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
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
			status, err := kc.getBackupStatus(u)
			if err != nil {
				kc.logger.Error()
			}
			err = kc.checkBackupStatus(status, u.GetNamespace(), u.GetName())
			if err != nil {
				kc.logger.Error("error checking backup status", err)
			} else {
				u.SetResourceVersion("123123")
				kc.logger.Info("backup is updated")
			}

		},
	})

	go informer.Run(kc.stopChan)

}

func (kc *KubernetesClient) Stop() {
	close(kc.stopChan)
}

func (kc *KubernetesClient) getBackupStatus(u *unstructured.Unstructured) (v1.BackupStatus, error) {
	var status v1.BackupStatus

	statusRaw, ok, err := unstructured.NestedFieldNoCopy(u.Object, "status")
	if err != nil {
		return status, err
	}
	if !ok {
		kc.logger.Error("status is missing")
		return status, errors.New("Status is Missing")
	}
	statusJSON, err := json.Marshal(statusRaw)
	if err != nil {
		return status, err
	}
	if err := json.Unmarshal(statusJSON, &status); err != nil {
		return status, err
	}
	return status, nil
}

func (kc *KubernetesClient) checkBackupStatus(status v1.BackupStatus, namespace string, name string) error {
	if status.Phase == "InProgress" {
		kc.logger.Infow("Backup is in InProgress mode nothing to do")
	} else if status.Phase == "Failed" {
		kc.logger.Errorw("Backup Failed ", "status", status.Phase, "Failed Reason ", status.FailureReason)
		messsage := "Backup Failed: " + name + " in namespace " + namespace + "\n for reason: " + status.FailureReason +
			"\n Please run `velero backup describe" + name + "` for more information"
		err := kc.kr.SendMessage(messsage, "Failed")
		if err != nil {
			kc.logger.Errorw("Cannot send message ", "error", err)
			return err
		}
	} else if status.Phase == "Completed" {
		kc.logger.Infow("Backup Completed", "status", status.Phase, "Items", status.Progress.ItemsBackedUp, "Time", status.CompletionTimestamp)
		msg := "Backup Completed " + name + " on namespace " + namespace + "\n Please run `velero backup describe " + name + "` for more information"
		err := kc.kr.SendMessage(msg, "Success")
		if err != nil {
			kc.logger.Errorw("Cannot send message ", "error", err)
			return err
		}
	} else {
		kc.logger.Infow("Backup is in different mode", "status", status.Phase)
	}
	return nil
}
