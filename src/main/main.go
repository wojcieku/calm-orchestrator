package main

import (
	"calm-orchestrator/src/controllers"
)

// TODO debug mode - wstrzymuje usuniecie zasobow w przypadku bledu

func main() {
	//configPath := flag.String("path", "", "Path to measurement configuration file")
	//flag.Parse()

	configMapController := controllers.NewCalmConfigMapController(controllers.GetClientSet())
	configMapController.Run()
	defer configMapController.Stop()

	// load/read config
	//configHandler := utils.MeasurementConfigHandler{}
	//config := configHandler.LoadConfigurationFromPath(*configPath)

	// TODO config bedzie z ConfigMapy - bedzie controller który przy każdym Create robi go measure(config)
	// controller musi mieć listę kanałów ? chyba nie musi

	// TODO ogarniecie sciezki do pliku z tokenami, to nizej zostaje do measure

	// go measure(config):
}
