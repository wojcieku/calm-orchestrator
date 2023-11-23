package main

import (
	"calm-orchestrator/src/commons"
	"calm-orchestrator/src/controller"
	"calm-orchestrator/src/utils"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

// TODO konfiguracja jako parametr wywolania albo cos + parametr kubeconfiga
const CONFIG_PATH = "sampleConfig.yaml"

func main() {
	// load/read config
	configHandler := utils.MeasurementConfigHandler{}
	config := configHandler.LoadConfigurationFromPath(CONFIG_PATH)

	// cluster names
	clientContextName := config.ClientSide
	serverContextName := config.ServerSide

	// prepare clients (kube config validation)
	serverSideClient := getDynamicClientWithContextName(serverContextName)
	clientSideClient := getDynamicClientWithContextName(clientContextName)

	// prepare LatencyMeasurements for Client and Server side
	serverSideLm := configHandler.ConfigToServerSideLatencyMeasurement(config)
	clientSideLm := configHandler.ConfigToClientSideLatencyMeasurement(config)

	// start controllers (prepare channels)
	serverSideStatusChan := make(chan string)
	clientSideStatusChan := make(chan string)

	serverSideController := controller.NewLatencyMeasurementController(serverSideClient, serverSideStatusChan)
	clientSideController := controller.NewLatencyMeasurementController(clientSideClient, clientSideStatusChan)

	go serverSideController.Run()
	defer serverSideController.Stop()

	go clientSideController.Run()
	defer clientSideController.Stop()

	// create Server side LM
	log.Info("Creating server side resources..")
	createLatencyMeasurement(serverSideClient, serverSideLm)

	// wait until completion of server side setup
serverStatusLoop:
	for status := range serverSideStatusChan {
		switch status {
		case commons.SUCCESS:
			{
				log.Info("Servers setup succeeded")
				break serverStatusLoop
			}
		case commons.FAILURE:
			{
				deleteLatencyMeasurement(serverSideClient, serverSideLm)
				log.Panic("Servers setup failed, deleting LatencyMeasurement in server side cluster..")
			}
		}
	}

	// create Client side LM
	log.Info("Creating client side resources..")
	createLatencyMeasurement(clientSideClient, clientSideLm)

	// wait until measurement completion
clientStatusLoop:
	for status := range clientSideStatusChan {
		switch status {
		case commons.SUCCESS:
			{
				log.Info("Clients setup and measurement succeeded")
				break clientStatusLoop
			}
		case commons.FAILURE:
			{
				deleteLatencyMeasurementsInBothClusters(serverSideClient, serverSideLm, clientSideClient, clientSideLm)
				log.Panic("Clients setup failed, deleting LatencyMeasurements in both clusters..")
			}
		}
	}
	fmt.Println("input text:")
	var w1, w2, w3 string
	_, _ = fmt.Scanln(&w1, &w2, &w3)

	// delete Client and Server side LMs
	deleteLatencyMeasurementsInBothClusters(serverSideClient, serverSideLm, clientSideClient, clientSideLm)
	log.Info("Measurement complete, all resources cleaned up")

	// completed, metrics?
}

func createLatencyMeasurement(client dynamic.Interface, lm commons.LatencyMeasurement) {
	serverSideLmMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&lm)
	if err != nil {
		log.Error("To unstructured conversion failed:", err)
	}
	_, err = client.Resource(commons.LatencyMeasurementResource).Namespace(commons.NAMESPACE).Create(context.Background(),
		&unstructured.Unstructured{Object: serverSideLmMap}, v1.CreateOptions{})

	if err != nil {
		deleteLatencyMeasurement(client, lm)
		log.Panic("Failed to create Server Side Latency Measurement: ", err)
	}
}

func deleteLatencyMeasurementsInBothClusters(serverSideClient dynamic.Interface, serverSideLm commons.LatencyMeasurement, clientSideClient dynamic.Interface, clientSideLm commons.LatencyMeasurement) {
	deleteLatencyMeasurement(serverSideClient, serverSideLm)
	deleteLatencyMeasurement(clientSideClient, clientSideLm)
}

func deleteLatencyMeasurement(client dynamic.Interface, lm commons.LatencyMeasurement) {
	err := client.Resource(commons.LatencyMeasurementResource).Namespace(commons.NAMESPACE).Delete(context.Background(), lm.Name, v1.DeleteOptions{})
	if err != nil {
		log.Errorf("LatencyMeasurement deletion in %s side cluster failed", lm.Spec.Side)
	}
}

func getDynamicClientWithContextName(contextName string) dynamic.Interface {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("error getting user home dir: %v\n", err)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	serverKubeConfig, err := buildConfigWithContextFromFlags(contextName, kubeConfigPath)

	if err != nil {
		log.Panicf("Failed to create k8s API client from context name: %v\n", err)
	}

	dynClient, err := dynamic.NewForConfig(serverKubeConfig)
	if err != nil {
		log.Panicf("error creating dynamic client: %v\n", err)
	}
	return dynClient
}

func buildConfigWithContextFromFlags(context string, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}
