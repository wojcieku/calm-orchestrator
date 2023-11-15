package main

import (
	controller2 "calm-orchestrator/src/controller"
	"calm-orchestrator/src/utils"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func main() {
	client, err := newClient()
	statusChannel := make(chan string)

	if err != nil {
		log.Error(err, "failed to create client")
	}
	controller, err := controller2.NewLatencyMeasurementController(client, statusChannel)
	if err != nil {
		log.Error(err, "failed to create LM controller")
	}
	log.Info("Starting controller")
	go controller.Run()

	// wait for success or failure of servers' deployment
	for status := range statusChannel {
		log.Info("Message received")
		switch status {
		case utils.SUCCESS:
			{
				log.Info("Setup succeeded")
			}
		case utils.FAILURE:
			{
				log.Error("Setup failed")
			}
		}
	}
	defer controller.Stop()
}

func newClient() (dynamic.Interface, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("error getting user home dir: %v\n", err)
		os.Exit(1)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		fmt.Printf("error getting Kubernetes config: %v\n", err)
		os.Exit(1)
	}

	dynClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		fmt.Printf("error creating dynamic client: %v\n", err)
		os.Exit(1)
	}

	return dynClient, nil
}

// odczytywanie
//client, err := newClient()
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
