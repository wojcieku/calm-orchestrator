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

	config := handler.LoadConfiguration(configFilePath)

	serverSideName := "klaster-serwerowy"
	if config.ServerSide != serverSideName {
		t.Error("Wrong ServerSide name parsed, expected:", serverSideName, "got:", config.ServerSide)
	}
	if len(config.Pairs) != 2 {
		t.Error("Quantity of measurement pairs invalid. Expected: 2, got:", len(config.Pairs))
	}
}

func TestMeasurementConfigHandler_ConfigToServerLatencyMeasurement(t *testing.T) {
	want := commons.LatencyMeasurement{
		ObjectMeta: metav1.ObjectMeta{Name: "measurement-2023-16-11"},
		Spec: commons.LatencyMeasurementSpec{Servers: []commons.Server{{
			Node:      "serverNode1",
			IpAddress: "10.10.10.10",
			Port:      1501,
		}, {
			Node:      "serverNode2",
			IpAddress: "20.20.20.20",
			Port:      2138,
		}}},
	}

	// TODO create MeasurementConfig with fixed data instead of calling LoadConfiguration()
	configFilePath := "sampleConfig.yaml"
	handler := MeasurementConfigHandler{}

	config := handler.LoadConfiguration(configFilePath)
	got := handler.ConfigToServerLatencyMeasurement(config)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ConfigToServerLatencyMeasurement() = %v, want %v", got, want)
	}
}
