package utils

import (
	"calm-orchestrator/src/commons"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
)

type MeasurementConfigHandler struct {
}

type MeasurementConfig struct {
	MeasurementID     string `yaml:"measurementID"`
	ServerSide        string `yaml:"serverSide"`
	ClientSide        string `yaml:"clientSide"`
	MetricsAggregator string `yaml:"metricsAggregator"`
	Pairs             []pair `yaml:"pairs"`
}

type pair struct {
	ServerNode string `yaml:"serverNode"`
	ClientNode string `yaml:"clientNode"`
	ServerIP   string `yaml:"serverIP"`
	ServerPort int    `yaml:"serverPort"`
	Interval   int    `yaml:"interval"`
	Duration   int    `yaml:"duration"`
}

func (m *MeasurementConfigHandler) LoadConfiguration(configFilePath string) MeasurementConfig {
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
	log.Info(fmt.Printf("Measurement config: %+v", measurementConfig))

	return measurementConfig
}

func (m *MeasurementConfigHandler) ConfigToServerLatencyMeasurement(config MeasurementConfig) commons.LatencyMeasurement {
	var lm commons.LatencyMeasurement
	// TODO dodac wersje API itd w typeMeta
	lm.Name = config.MeasurementID

	for _, p := range config.Pairs {
		server := commons.Server{
			Node:      p.ServerNode,
			IpAddress: p.ServerIP,
			Port:      p.ServerPort,
		}
		lm.Spec.Servers = append(lm.Spec.Servers, server)
	}
	return lm
}

//func ConfigToClientSideLatencyMeasurement(config MeasurementConfig) (commons.LatencyMeasurement, error) {
//
//}