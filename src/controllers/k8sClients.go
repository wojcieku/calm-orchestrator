package controllers

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

// TODO ogarniecie sciezek w secrecie itd
func GetClientSet() *kubernetes.Clientset {
	userHomeDir, _ := os.UserHomeDir()
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Panicf("Could not obtain kube config file")
		return nil
	}
	set, _ := kubernetes.NewForConfig(restConfig)
	return set
}

func getDynamicClientWithContextName(contextName string) dynamic.Interface {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("error getting user home dir: %v\n", err)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config-outside-clusters")
	serverKubeConfig, err := buildConfigWithContextFromFlags(contextName, kubeConfigPath)

	if err != nil {
		log.Panicf("Failed to create k8s API client from context name: %v\n", err)
	}

	dynClient, err := dynamic.NewForConfig(serverKubeConfig)
	if err != nil {
		log.Panicf("error creating dynamic client: %v\n", err)
	}
	return dynClient
}

func buildConfigWithContextFromFlags(context string, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}
