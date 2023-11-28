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
	informer        cache.SharedIndexInformer
	stopper         chan struct{}
	queue           workqueue.RateLimitingInterface
	status          chan string
	measurementName string
}

func NewLatencyMeasurementController(client dynamic.Interface, statusChan chan string, measurementName string) *LatencyMeasurementController {
	dynInformer := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, commons.NAMESPACE, nil)
	informer := dynInformer.ForResource(commons.LatencyMeasurementResource).Informer()
	stopper := make(chan struct{})

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: getUpdateFunc(queue, measurementName),
	})

	return &LatencyMeasurementController{
		informer:        informer,
		queue:           queue,
		stopper:         stopper,
		status:          statusChan,
		measurementName: measurementName,
	}
}

func getUpdateFunc(queue workqueue.RateLimitingInterface, measurementName string) func(oldObj interface{}, newObj interface{}) {
	return func(oldObj, newObj interface{}) {
		converter := runtime.DefaultUnstructuredConverter
		unstructured, err := converter.ToUnstructured(newObj)
		if err != nil {
			log.Error("could not convert: ", err.Error())
		}
		var lm commons.LatencyMeasurement

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &lm)
		if err != nil {
			log.Error("Event conversion to LatencyMeasurement failed")
			return
		}
		// filter events only for this measurement
		if lm.Name != measurementName {
			return
		}
		if lm.Status.State == commons.SUCCESS {
			queue.Add(lm.Status.State)
		} else if lm.Status.State == commons.FAILURE {
			log.Errorf("Received failure status: %s", lm.Status.Details)
			queue.Add(lm.Status.State)
		}
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
	switch key {
	case commons.SUCCESS:
		l.status <- commons.SUCCESS
	case commons.FAILURE:
		// TODO send details from status State and Details
		l.status <- commons.FAILURE
	}
	return nil
}
