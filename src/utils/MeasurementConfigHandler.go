package utils

import (
	"calm-orchestrator/src/commons"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
)

type MeasurementConfigHandler struct {
}

type MeasurementConfig struct {
	MeasurementID            string `yaml:"measurementID"`
	ServerSideClusterName    string `yaml:"serverSideClusterName"`
	ClientSideClusterName    string `yaml:"clientSideClusterName"`
	MetricsAggregatorAddress string `yaml:"metricsAggregatorAddress"`
	Pairs                    []Pair `yaml:"pairs"`
}

type Pair struct {
	ServerNodeName string `yaml:"serverNodeName"`
	ClientNodeName string `yaml:"clientNodeName"`
	ServerIP       string `yaml:"serverIP"`
	ServerPort     int    `yaml:"serverPort"`
	Interval       int    `yaml:"interval"`
	Duration       int    `yaml:"duration"`
}

func (m *MeasurementConfigHandler) LoadConfigurationFromPath(configFilePath string) MeasurementConfig {
	f, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Error("Failed to read file: " + err.Error())
		os.Exit(1)
	}
	var measurementConfig MeasurementConfig
	err = yaml.Unmarshal(f, &measurementConfig)

	if err != nil {
		log.Error("Failed to unmarshal configuration: " + err.Error())
		os.Exit(1)
	}
	return measurementConfig
}

func (m *MeasurementConfigHandler) ConfigToServerSideLatencyMeasurement(config MeasurementConfig) commons.LatencyMeasurement {
	lm := getDefaultLatencyMeasurement(config)
	lm.Spec.Side = commons.SERVER_SIDE
	for _, p := range config.Pairs {
		server := commons.Server{
			ServerNodeName: p.ServerNodeName,
			ClientNodeName: p.ClientNodeName,
			ServerIP:       p.ServerIP,
			ServerPort:     p.ServerPort,
		}
		lm.Spec.Servers = append(lm.Spec.Servers, server)
	}
	return lm
}

func (m *MeasurementConfigHandler) ConfigToClientSideLatencyMeasurement(config MeasurementConfig) commons.LatencyMeasurement {
	lm := getDefaultLatencyMeasurement(config)
	lm.Spec.Side = commons.CLIENT_SIDE
	for _, p := range config.Pairs {
		client := commons.Client{
			ServerIP:                 p.ServerIP,
			ServerPort:               p.ServerPort,
			Interval:                 p.Interval,
			Duration:                 p.Duration,
			MetricsAggregatorAddress: config.MetricsAggregatorAddress,
			ClientNodeName:           p.ClientNodeName,
			ServerNodeName:           p.ServerNodeName,
			ClientSideClusterName:    config.ClientSideClusterName,
			ServerSideClusterName:    config.ServerSideClusterName,
		}
		lm.Spec.Clients = append(lm.Spec.Clients, client)
	}
	return lm
}

func getDefaultLatencyMeasurement(config MeasurementConfig) commons.LatencyMeasurement {
	return commons.LatencyMeasurement{
		TypeMeta: metav1.TypeMeta{
			Kind:       commons.KIND,
			APIVersion: commons.API_GROUP_WITH_VERSION,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.MeasurementID,
			Namespace: commons.NAMESPACE,
		},
	}
}
