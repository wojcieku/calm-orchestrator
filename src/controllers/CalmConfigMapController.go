package controllers

import (
	"calm-orchestrator/src/commons"
	"calm-orchestrator/src/utils"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/yaml"
	"time"
)

type CalmConfigMapController struct {
	informer cache.SharedIndexInformer
	stopper  chan struct{}
}

func NewCalmConfigMapController(client kubernetes.Interface) *CalmConfigMapController {
	factory := informers.NewSharedInformerFactoryWithOptions(client, 1*time.Second, informers.WithNamespace(commons.NAMESPACE))
	informer := factory.Core().V1().ConfigMaps().Informer()
	stopper := make(chan struct{})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: getConfigMapAddFunc(),
	})

	return &CalmConfigMapController{
		informer: informer,
		stopper:  stopper,
	}
}

func getConfigMapAddFunc() func(obj interface{}) {
	return func(obj interface{}) {
		configMap := obj.(*v1.ConfigMap)
		for _, data := range configMap.Data {
			var config utils.MeasurementConfig
			err := yaml.Unmarshal([]byte(data), &config)
			if err != nil {
				log.Errorf("Error in json configmap into struct conversion: %s", err)
				return
			}
			go launchMeasurement(config)
		}
	}
}

func (c *CalmConfigMapController) Stop() {
	close(c.stopper)
}

func (c *CalmConfigMapController) Run() {
	defer utilruntime.HandleCrash()
	c.informer.Run(c.stopper)

}
