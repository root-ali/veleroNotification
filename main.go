package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/root-ali/velero-reporter/notifier"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

func main() {
	// Allow passing a kubeconfig for local development.
	var kubeconfig string
	flag.StringVar(&kubeconfig, "kubeconfig", "/home/root-ali/.kube/config", "absolute path to the kubeconfig file")
	flag.Parse()

	// Build the config.
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig if not running in-cluster.
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %v", err)
		}
	}

	// Create a dynamic Kubernetes client.
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating dynamic client: %v", err)
	}

	// Define the Velero Backup resource.
	backupRes := schema.GroupVersionResource{
		Group:    "velero.io",
		Version:  "v1",
		Resource: "backups",
	}

	// Create a ListWatch for all Backup objects.
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return dynamicClient.Resource(backupRes).Namespace(metav1.NamespaceAll).List(context.Background(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return dynamicClient.Resource(backupRes).Namespace(metav1.NamespaceAll).Watch(context.Background(), options)
		},
	}

	// Create an informer on the Backup CRD.
	informer := cache.NewSharedInformer(
		lw,
		&unstructured.Unstructured{},
		0, // resync period; 0 means no periodic resync.
	)

	// Set up event handlers.
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			msg := fmt.Sprintf("Backup added: %s/%s", u.GetNamespace(), u.GetName())
			log.Println(msg)
			if err := notifier.Notify(msg); err != nil {
				log.Printf("Error sending notification: %v", err)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			u := newObj.(*unstructured.Unstructured)
			msg := fmt.Sprintf("Backup updated: %s/%s", u.GetNamespace(), u.GetName())
			log.Println(msg)
			if err := notifier.Notify(msg); err != nil {
				log.Printf("Error sending notification: %v", err)
			}
		},
		DeleteFunc: func(obj interface{}) {
			u, ok := obj.(*unstructured.Unstructured)
			if !ok {
				// Handle tombstone events.
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					log.Printf("Could not get object from tombstone %+v", obj)
					return
				}
				u, ok = tombstone.Obj.(*unstructured.Unstructured)
				if !ok {
					log.Printf("Tombstone contained object that is not Unstructured: %+v", obj)
					return
				}
			}
			msg := fmt.Sprintf("Backup deleted: %s/%s", u.GetNamespace(), u.GetName())
			log.Println(msg)
			if err := notifier.Notify(msg); err != nil {
				log.Printf("Error sending notification: %v", err)
			}
		},
	})

	// Start the informer.
	stopCh := make(chan struct{})
	defer close(stopCh)
	go informer.Run(stopCh)

	// Wait indefinitely.
	select {}
}
