package controller

import (
	"calm-orchestrator/src/commons"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const maxRetries = 3

type LatencyMeasurementController struct {
	informer cache.SharedIndexInformer
	stopper  chan struct{}
	queue    workqueue.RateLimitingInterface
	status   chan string
}

func NewLatencyMeasurementController(client dynamic.Interface, statusChan chan string) *LatencyMeasurementController {
	// TODO for namespace
	dynInformer := dynamicinformer.NewDynamicSharedInformerFactory(client, 0)
	informer := dynInformer.ForResource(commons.LatencyMeasurementResource).Informer()
	stopper := make(chan struct{})

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// TODO obsluga dodania i usuniecia?
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
		UpdateFunc: getUpdateFunc(queue),
	})

	return &LatencyMeasurementController{
		informer: informer,
		queue:    queue,
		stopper:  stopper,
		status:   statusChan,
	}
}

func getUpdateFunc(queue workqueue.RateLimitingInterface) func(oldObj interface{}, newObj interface{}) {
	return func(oldObj, newObj interface{}) {
		converter := runtime.DefaultUnstructuredConverter
		unstructured, err := converter.ToUnstructured(newObj)
		if err != nil {
			log.Error("could not convert: ", err.Error())
		}
		var lm commons.LatencyMeasurement

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &lm)
		if err != nil {
			log.Info("Event conversion to LatencyMeasurement failed")
			return
		}
		queue.Add(lm.Status.State)
	}
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
		switch {
		case err == nil:
			l.queue.Forget(key)
		case l.queue.NumRequeues(key) < maxRetries:
			l.queue.AddRateLimited(key)
		default:
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
	case commons.SUCCESS:
		// TODO send success
		log.Info("Update event logic executed")
		l.status <- commons.SUCCESS
	case "delete":
		log.Info("Delete event logic executed")
	}
	return nil
}
