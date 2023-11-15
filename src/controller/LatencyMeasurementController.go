package controller

import (
	"calm-orchestrator/src/utils"
	"fmt"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

type Server struct {
	Node      string `json:"node,omitempty"`
	IpAddress string `json:"ip_address,omitempty"`
	Port      int    `json:"port,omitempty"`
}

type Client struct {
	Node              string `json:"node,omitempty"`
	IpAddress         string `json:"ip_address,omitempty"`
	Port              int    `json:"port,omitempty"`
	Interval          int    `json:"interval,omitempty"`
	Duration          int    `json:"duration,omitempty"`
	MetricsAggregator string `json:"metricsAggregator,omitempty"`
}

// LatencyMeasurementSpec defines the desired state of LatencyMeasurement
type LatencyMeasurementSpec struct {
	Side    string   `json:"side,omitempty"`
	Servers []Server `json:"servers,omitempty"`
	Clients []Client `json:"clients,omitempty"`
}

// LatencyMeasurementStatus defines the observed state of LatencyMeasurement
type LatencyMeasurementStatus struct {
	State   string `json:"state,omitempty"`
	Details string `json:"details,omitempty"`
}

type LatencyMeasurement struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LatencyMeasurementSpec   `json:"spec,omitempty"`
	Status LatencyMeasurementStatus `json:"status,omitempty"`
}

type LatencyMeasurementController struct {
	informer cache.SharedIndexInformer
	stopper  chan struct{}
	queue    workqueue.RateLimitingInterface
	status   chan string
}

func NewLatencyMeasurementController(client dynamic.Interface, statusChan chan string) (*LatencyMeasurementController, error) {
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
			converter := runtime.DefaultUnstructuredConverter
			unstructured, err := converter.ToUnstructured(newObj)
			if err != nil {
				log.Error("could not convert: ", err.Error())
			}
			var lm LatencyMeasurement

			err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured, &lm)
			if err != nil {
				log.Info("Event conversion to LatencyMeasurement failed")
				return
			}
			queue.Add(lm.Status.State)

		},
	})

	return &LatencyMeasurementController{
		informer: informer,
		queue:    queue,
		stopper:  stopper,
		status:   statusChan,
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
	case utils.SUCCESS:
		// TODO send success
		log.Info("Update event logic executed")
		l.status <- utils.SUCCESS
	case "delete":
		log.Info("Delete event logic executed")
	}
	return nil
}
