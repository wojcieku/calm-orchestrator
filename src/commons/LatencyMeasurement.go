package commons

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	SUCCESS = "Success"
	FAILURE = "Failure"
)

var LatencyMeasurementResource = schema.GroupVersionResource{
	Group:    "measurement.calm.com",
	Version:  "v1alpha1",
	Resource: "latencymeasurements",
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
	Node      string `json:"node,omitempty"`
	IpAddress string `json:"ip_address,omitempty"`
	Port      int    `json:"port,omitempty"`
}

type Client struct {
	Node              string `json:"node,omitempty"`
	IpAddress         string `json:"ip_address,omitempty"`
	Port              int    `json:"port,omitempty"`
	Interval          int    `json:"interval,omitempty"`
	Duration          int    `json:"duration,omitempty"`
	MetricsAggregator string `json:"metricsAggregator,omitempty"`
}
