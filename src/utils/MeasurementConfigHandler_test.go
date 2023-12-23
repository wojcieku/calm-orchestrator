package utils

import (
	"calm-orchestrator/src/commons"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestMeasurementConfigHandler_LoadConfiguration(t *testing.T) {
	configFilePath := "sampleConfig.yaml"
	handler := MeasurementConfigHandler{}

	config := handler.LoadConfigurationFromPath(configFilePath)

	serverSideName := "klaster-serwerowy"
	if config.ServerSideClusterName != serverSideName {
		t.Error("Wrong ServerSideClusterName name parsed, expected:", serverSideName, "got:", config.ServerSideClusterName)
	}
	if len(config.Pairs) != 2 {
		t.Error("Quantity of measurement pairs invalid. Expected: 2, got:", len(config.Pairs))
	}
}

func TestMeasurementConfigHandler_ConfigToServerLatencyMeasurement(t *testing.T) {
	want := commons.LatencyMeasurement{
		TypeMeta: metav1.TypeMeta{
			Kind:       commons.KIND,
			APIVersion: commons.API_GROUP_WITH_VERSION,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "measurement-2023-16-11",
			Namespace: commons.NAMESPACE,
		},
		Spec: commons.LatencyMeasurementSpec{Servers: []commons.Server{{
			ServerNodeName:  "serverNode1",
			ServerIPAddress: "10.10.10.10",
			ServerPort:      1501,
		}, {
			ServerNodeName:  "serverNode2",
			ServerIPAddress: "20.20.20.20",
			ServerPort:      2138,
		}}},
	}

	// TODO create MeasurementConfig with fixed data instead of calling LoadConfigurationFromPath()
	configFilePath := "../../sampleConfig.yaml"
	handler := MeasurementConfigHandler{}

	config := handler.LoadConfigurationFromPath(configFilePath)
	got := handler.ConfigToServerSideLatencyMeasurement(config)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ConfigToServerSideLatencyMeasurement() = %v, want %v", got, want)
	}
}
