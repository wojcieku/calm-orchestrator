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
const CONFIG_PATH = "../../sampleConfig.yaml"

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
	serverSideLmMap, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(serverSideLm)
	_, err := serverSideClient.Resource(commons.LatencyMeasurementResource).Create(context.Background(),
		&unstructured.Unstructured{Object: serverSideLmMap}, v1.CreateOptions{})

	if err != nil {
		log.Fatal("Failed to create Server Side Latency Measurement")
	}

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
				log.Error("Servers setup failed")
				// TODO custom resources clean up
			}
		}
	}

	// create Client side LM
	clientSideLmMap, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(clientSideLm)
	_, err = clientSideClient.Resource(commons.LatencyMeasurementResource).Create(context.Background(),
		&unstructured.Unstructured{Object: clientSideLmMap}, v1.CreateOptions{})

	if err != nil {
		log.Fatal("Failed to create Client Side Latency Measurement")
	}

clientStatusLoop:
	for status := range serverSideStatusChan {
		switch status {
		case commons.SUCCESS:
			{
				log.Info("Clients setup and measurement succeeded")
				break clientStatusLoop
			}
		case commons.FAILURE:
			{
				log.Error("Clients setup failed, measurement failed")
				// TODO custom resources clean up
			}
		}
	}

	// delete Client and Server side LMs

	// completed, metrics?
}

func getDynamicClientWithContextName(contextName string) dynamic.Interface {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("error getting user home dir: %v\n", err)
		os.Exit(1)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)
	serverKubeConfig, err := buildConfigWithContextFromFlags(contextName, kubeConfigPath)

	dynClient, err := dynamic.NewForConfig(serverKubeConfig)
	if err != nil {
		fmt.Printf("error creating dynamic client: %v\n", err)
		os.Exit(1)
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
