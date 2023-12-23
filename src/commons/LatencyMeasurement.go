package commons

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	SUCCESS                = "Success"
	FAILURE                = "Failure"
	NAMESPACE              = "calm-operator-system"
	KIND                   = "LatencyMeasurement"
	API_GROUP              = "measurement.calm.com"
	API_VERSION            = "v1alpha1"
	API_RESOURCE           = "latencymeasurements"
	API_GROUP_WITH_VERSION = API_GROUP + "/" + API_VERSION
	SERVER_SIDE            = "server"
	CLIENT_SIDE            = "client"
)

var LatencyMeasurementResource = schema.GroupVersionResource{
	Group:    API_GROUP,
	Version:  API_VERSION,
	Resource: API_RESOURCE,
}

type LatencyMeasurement struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LatencyMeasurementSpec   `json:"spec,omitempty"`
	Status LatencyMeasurementStatus `json:"status,omitempty"`
}

// LatencyMeasurementSpec defines the desired state of LatencyMeasurement
type LatencyMeasurementSpec struct {
	Side    string   `json:"side,omitempty"`
	Servers []Server `json:"servers,omitempty"`
	Clients []Client `json:"clients,omitempty"`
}

// LatencyMeasurementStatus defines the observed state of LatencyMeasurement
type LatencyMeasurementStatus struct {
	State   string `json:"state,omitempty"`
	Details string `json:"details,omitempty"`
}

type Server struct {
	ServerNodeName  string `json:"serverNodeName,omitempty"`
	ClientNodeName  string `json:"clientNodeName,omitempty"`
	ServerIPAddress string `json:"serverIPAddress,omitempty"`
	ServerPort      int    `json:"serverPort,omitempty"`
}

type Client struct {
	IPAddress            string `json:"ipAddress,omitempty"`
	Port                 int    `json:"port,omitempty"`
	Interval             int    `json:"interval,omitempty"`
	Duration             int    `json:"duration,omitempty"`
	MetricsAggregatorURL string `json:"metricsAggregatorURL,omitempty"`
	ClientNodeName       string `json:"clientNodeName,omitempty"`
	ServerNodeName       string `json:"serverNodeName,omitempty"`
	ClientClusterName    string `json:"clientClusterName,omitempty"`
	ServerClusterName    string `json:"serverClusterName,omitempty"`
}
