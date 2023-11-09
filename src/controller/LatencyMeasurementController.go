package controller

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const maxRetries = 3

var latencyMeasurementResource = schema.GroupVersionResource{
	Group:    "measurement.calm.com",
	Version:  "v1alpha1",
	Resource: "latencymeasurements",
}

type LatencyMeasurementController struct {
	informer cache.SharedIndexInformer
	stopper  chan struct{}
	queue    workqueue.RateLimitingInterface
}

func NewLatencyMeasurementController(client dynamic.Interface) (*LatencyMeasurementController, error) {
	// TODO for namespace
	dynInformer := dynamicinformer.NewDynamicSharedInformerFactory(client, 0)
	informer := dynInformer.ForResource(latencyMeasurementResource).Informer()
	stopper := make(chan struct{})

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		AddFunc: func(obj interface{}) {
			log.Info("AddFunc object: ", obj)
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				// tu mozna dodac co sie chce chyba do kolejki
				queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			log.Info("Update old object:", oldObj)
			log.Info("Update new object:", newObj)
			jsonStr, ok := newObj.(*string)
			if ok {
				str := *jsonStr
				x := map[string]string{}
				err := json.Unmarshal([]byte(str), &x)
				if err != nil {
					log.Error("Error unmarshalling:", err.Error())
				}
			} else {
				log.Info("newObj is not string")
			}
		},
	})

	return &LatencyMeasurementController{
		informer: informer,
		queue:    queue,
		stopper:  stopper,
	}, nil
}

func (l *LatencyMeasurementController) Stop() {
	close(l.stopper)
}

func (l *LatencyMeasurementController) Run() {
	defer utilruntime.HandleCrash()

	defer l.queue.ShutDown()

	go l.informer.Run(l.stopper)

	// wait for the caches to synchronize before starting the worker
	if !cache.WaitForCacheSync(l.stopper, l.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	// runWorker will loop until some problem happens. The wait.Until will then restart the worker after one second
	wait.Until(l.runWorker, time.Second, l.stopper)
}

func (l *LatencyMeasurementController) runWorker() {
	for {
		key, quit := l.queue.Get()
		if quit {
			return
		}

		err := l.processItem(key.(string))

		if err == nil {
			l.queue.Forget(key)
		} else if l.queue.NumRequeues(key) < maxRetries {
			l.queue.AddRateLimited(key)
		} else {
			l.queue.Forget(key)
			utilruntime.HandleError(err)
		}

		l.queue.Done(key)
	}
}

func (l *LatencyMeasurementController) processItem(key string) error {
	log.Info("process Item key: " + key)
	switch key {
	case "create":
		log.Info("Create event logic executed")
	case "update":
		log.Info("Update event logic executed")
	case "delete":
		log.Info("Delete event logic executed")
	}
	return nil
}
