package controllers

import (
	"calm-orchestrator/src/commons"
	"calm-orchestrator/src/utils"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

type CalmConfigMapController struct {
	informer cache.SharedIndexInformer
	stopper  chan struct{}
	queue    workqueue.RateLimitingInterface
}

// TODO ewentualnie mozna zrobic liste factories

func NewCalmConfigMapController(client *kubernetes.Clientset) *CalmConfigMapController {
	factory := informers.NewSharedInformerFactoryWithOptions(client, 5, informers.WithNamespace(commons.NAMESPACE))
	informer := factory.Core().V1().ConfigMaps().Informer()
	stopper := make(chan struct{})

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: getConfigMapAddFunc(queue),
	})

	return &CalmConfigMapController{
		informer: informer,
		queue:    queue,
		stopper:  stopper,
	}
}

func getConfigMapAddFunc(queue workqueue.RateLimitingInterface) func(obj interface{}) {
	return func(obj interface{}) {
		configMap := obj.(*v1.ConfigMap)
		jsonData, err := json.Marshal(configMap.Data)
		if err != nil {
			log.Errorf("Error in configmap into json conversion: %s", err)
			return
		}
		var config utils.MeasurementConfig
		err = json.Unmarshal(jsonData, &config)
		if err != nil {
			log.Errorf("Error in json configmap into struct conversion: %s", err)
			return
		}
		queue.Add(config)
		go launchMeasurement(config)
	}
}

func (c *CalmConfigMapController) Stop() {
	close(c.stopper)
}

func (c *CalmConfigMapController) Run() {
	defer utilruntime.HandleCrash()

	defer c.queue.ShutDown()

	go c.informer.Run(c.stopper)

	// wait for the caches to synchronize before starting the worker
	if !cache.WaitForCacheSync(c.stopper, c.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	// runWorker will loop until some problem happens. The wait.Until will then restart the worker after one second
	wait.Until(c.runWorker, time.Second, c.stopper)
}

func (c *CalmConfigMapController) runWorker() {
	for {
		key, quit := c.queue.Get()
		if quit {
			return
		}
		err := c.processItem(key)
		switch {
		case err == nil:
			c.queue.Forget(key)
		case c.queue.NumRequeues(key) < maxRetries:
			c.queue.AddRateLimited(key)
		default:
			c.queue.Forget(key)
			utilruntime.HandleError(err)
		}
		c.queue.Done(key)
	}
}

func (c *CalmConfigMapController) processItem(item interface{}) error {
	config, ok := item.(utils.MeasurementConfig)
	if !ok {
		err := errors.New("could not map item into measurement config")
		log.Error(err)
		return err
	}
	go launchMeasurement(config)
	return nil
}
