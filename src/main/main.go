package main

import (
	"calm-orchestrator/src/utils"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

// TODO konfiguracja jako parametr wywolania albo cos
const CONFIG_PATH = "../../sampleConfig.yaml"

func main() {
	// load/read config
	configHandler := utils.MeasurementConfigHandler{}
	config := configHandler.LoadConfigurationFromPath(CONFIG_PATH)

	clientContextName := config.ClientSide
	serverContextName := config.ServerSide

	// prepare clients (kube config validation)

	// prepare LatencyMeasurements for Client and Server side
	serverSideLm := configHandler.ConfigToServerSideLatencyMeasurement(config)
	clientSideLm := configHandler.ConfigToClientSideLatencyMeasurement(config)

	// start controllers (prepare channels)
	serverSideStatusChan := make(chan string)
	clientSideStatusChan := make(chan string)

	// create Server side LM

	// create Client side LM

	// delete Client and Server side LMs

	// completed, metrics?

	clients, err := getClients()
	//statusChannel := make(chan string)

	if err != nil {
		log.Error(err, "failed to create client")
	}
	serverSideClient := clients[0]
	clientSideClient := clients[1]

	pods, err := serverSideClient.Resource(gvr).Namespace("").List(context.Background(), v1.ListOptions{})

	if err != nil {
		fmt.Printf("error getting pods: %v\n", err)
		os.Exit(1)
	}

	log.Info("Listing for server side")
	for _, pod := range pods.Items {
		log.Info(
			"Name: %s\n",
			pod.Object["metadata"].(map[string]interface{})["name"],
		)
	}

	pods, err = clientSideClient.Resource(gvr).Namespace("").List(context.Background(), v1.ListOptions{})

	if err != nil {
		fmt.Printf("error getting pods: %v\n", err)
		os.Exit(1)
	}

	log.Info("Listing for client side")
	for _, pod := range pods.Items {
		log.Info(
			"Name: %s\n",
			pod.Object["metadata"].(map[string]interface{})["name"],
		)
	}
	//controller, err := controller2.NewLatencyMeasurementController(client, statusChannel)
	//if err != nil {
	//	log.Error(err, "failed to create LM controller")
	//}
	//log.Info("Starting controller")
	//go controller.Run()
	//
	//// wait for success or failure of servers' deployment
	//for status := range statusChannel {
	//	log.Info("Message received")
	//	switch status {
	//	case utils.SUCCESS:
	//		{
	//			log.Info("Setup succeeded")
	//		}
	//	case utils.FAILURE:
	//		{
	//			log.Error("Setup failed")
	//		}
	//	}
	//}
	//defer controller.Stop()
}

func getClients() ([]dynamic.Interface, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("error getting user home dir: %v\n", err)
		os.Exit(1)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

	clientCluster := "kind-client-side"
	serverCluster := "kind-server-side"

	serverKubeConfig, err := buildConfigWithContextFromFlags(serverCluster, kubeConfigPath)
	//kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		fmt.Printf("error getting Kubernetes config: %v\n", err)
		os.Exit(1)
	}
	clientKubeConfig, err := buildConfigWithContextFromFlags(clientCluster, kubeConfigPath)
	//kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		fmt.Printf("error getting Kubernetes config: %v\n", err)
		os.Exit(1)
	}

	serverSideDynClient, err := dynamic.NewForConfig(serverKubeConfig)
	if err != nil {
		fmt.Printf("error creating dynamic client: %v\n", err)
		os.Exit(1)
	}
	clientSideDynClient, err := dynamic.NewForConfig(clientKubeConfig)
	if err != nil {
		fmt.Printf("error creating dynamic client: %v\n", err)
		os.Exit(1)
	}

	return []dynamic.Interface{serverSideDynClient, clientSideDynClient}, nil
}

func getDynamicClientForContextName(contextName string) dynamic.Interface {
	// TODO to samo co wyzej tylko dla jednego
}

func buildConfigWithContextFromFlags(context string, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

// odczytywanie
//client, err := getClients()
//
//gvr := schema.GroupVersionResource{
//	Group:    "measurement.calm.com",
//	Version:  "v1alpha1",
//	Resource: "latencymeasurements",
//}
//
//pods, err := client.Resource(gvr).Namespace("calm-operator-system").List(context.Background(), v1.ListOptions{})
//if err != nil {
//	fmt.Printf("error getting pods: %v\n", err)
//	os.Exit(1)
//}
//
//for _, pod := range pods.Items {
//	fmt.Printf(
//		"Name: %s\n",
//		pod.Object["metadata"].(map[string]interface{})["name"],
//	)
//}
