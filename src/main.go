package main

import (
	"calm-orchestrator/src/controllers"
)

// TODO debug mode - wstrzymuje usuniecie zasobow w przypadku bledu

func main() {
	configMapController := controllers.NewCalmConfigMapController(controllers.GetClientSet())
	configMapController.Run()
	defer configMapController.Stop()
}
