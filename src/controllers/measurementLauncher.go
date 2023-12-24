package controllers

import (
	"calm-orchestrator/src/commons"
	"calm-orchestrator/src/utils"
	"context"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

func launchMeasurement(config utils.MeasurementConfig) {
	configHandler := utils.MeasurementConfigHandler{}

	// cluster names
	clientContextName := config.ClientSideClusterName
	serverContextName := config.ServerSideClusterName

	// prepare clients (kube config validation)
	serverSideClient := getDynamicClientWithContextName(serverContextName)
	clientSideClient := getDynamicClientWithContextName(clientContextName)

	// prepare LatencyMeasurements for Client and Server side
	serverSideLm := configHandler.ConfigToServerSideLatencyMeasurement(config)
	clientSideLm := configHandler.ConfigToClientSideLatencyMeasurement(config)

	// start controllers (prepare channels)
	serverSideStatusChan := make(chan string)
	clientSideStatusChan := make(chan string)

	serverSideController := NewLatencyMeasurementController(serverSideClient, serverSideStatusChan, config.MeasurementID)
	clientSideController := NewLatencyMeasurementController(clientSideClient, clientSideStatusChan, config.MeasurementID)

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
				//fmt.Println("Resources will be deleted, check logs or press any key to proceed")
				//_, _ = fmt.Scanln()
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
				//fmt.Println("Resources will be deleted, check logs or press any key to proceed")
				//_, _ = fmt.Scanln()
				deleteLatencyMeasurementsInBothClusters(serverSideClient, serverSideLm, clientSideClient, clientSideLm)
				log.Panic("Clients setup failed, deleting LatencyMeasurements in both clusters..")
			}
		}
	}

	// delete Client and Server side LMs
	deleteLatencyMeasurementsInBothClusters(serverSideClient, serverSideLm, clientSideClient, clientSideLm)
	log.Infof("Measurement %s complete, all resources cleaned up", config.MeasurementID)
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
