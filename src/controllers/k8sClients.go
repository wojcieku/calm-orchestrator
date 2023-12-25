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

// TODO ogarniecie sciezek w secrecie itd - trzeba zrobić Secret z keys: config i config-outside-clusers a w deploymencie dać:
//  volumeMounts:
//          - mountPath: "/.kube"
//            name: cluster-configs-mount
//            readOnly: true
//      volumes:
//        - name: cluster-configs-mount
//          secret:
//            secretName: cluster-configs

func GetClientSet() *kubernetes.Clientset {
	kubeConfigPath := filepath.Join("/var/management", "config")
	configFileContent, err := os.ReadFile(kubeConfigPath)
	if err != nil {
		log.Errorf("Could not read kube config file: %s", err)
		return nil
	}
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(configFileContent)
	if err != nil {
		log.Errorf("Could not obtain kube config file: %s", err)
		return nil
	}
	set, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Errorf("Could not create client set: %s", err)
	}
	return set
}

func getDynamicClientWithContextName(contextName string) dynamic.Interface {
	kubeConfigPath := filepath.Join("/var/outside", "config-outside-clusters")
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
