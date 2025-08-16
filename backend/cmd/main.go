package main

import (
	postgres "backend/database"
	mqttClient "backend/pkg/mqtt"
	"backend/pkg/utils"
	"backend/pkg/websockets"
	"log"
)

// var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
// 	log.Println("Connected")
// }

// var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
// 	log.Printf("Connect lost: %v", err)
// }

func main() {
	// Load environment variables from .env file
	utils.LoadEnv()

	// Initialize database connection
	if err := postgres.InitDb(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Websocket
	wsHub := websockets.StartWebsocketServer()

	// Start API server
	StartAPIServer()

	certs_folder := utils.GetEnv("MQTT_CERTS_PATH", "../certs")
	cert_bundle_name := utils.GetEnv("MQTT_CERT_BUNDLE_NAME", "bundle.pem")

	// MQTT client configuration
	mqttConfig := mqttClient.MqttConfig{
		Broker:   utils.GetEnv("MQTT_BROKER", "mqtt://localhost:1883"),
		ClientId: utils.GetEnv("MQTT_CLIENT_ID", "home-security-backend"),
		Username: utils.GetEnv("MQTT_USERNAME", ""),
		Password: utils.GetEnv("MQTT_PASSWORD", ""),
		CAPath:   utils.GetEnv("MQTT_CAFILE", certs_folder+"/"+cert_bundle_name),
		CertPath: utils.GetEnv("MQTT_CERT_PATH", certs_folder+"/client.crt"),
		KeyPath:  utils.GetEnv("MQTT_KEY_PATH", certs_folder+"/client.key"),
	}
	log.Printf("Cert file paths: %s, %s, %s\n", mqttConfig.CAPath, mqttConfig.CertPath, mqttConfig.KeyPath)
	mqttClient := mqttClient.NewMqttClient(mqttConfig, wsHub)
	err := mqttClient.Connect()
	if err != nil {
		log.Printf("Failed to connect to MQTT broker: %v", err)
		log.Println("Application will continue running without MQTT connectivity")
	}

	// Keep the program running
	select {}
}
