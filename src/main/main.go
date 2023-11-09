package main

import (
	controller2 "calm-orchestrator/src/controller"
	"fmt"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = logf.Log.WithName("main")

func main() {
	client, err := newClient()
	if err != nil {
		logger.Error(err, "failed to create client")
	}
	controller, err := controller2.NewLatencyMeasurementController(client)
	if err != nil {
		logger.Error(err, "failed to create LM controller")
	}
	controller.Run()
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
