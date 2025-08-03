package main

import (
	postgres "backend/database"
	"backend/database/models"
	"backend/database/services"
	"backend/pkg/utils"
	"backend/pkg/websockets"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}

func main() {
	// Load environment variables from .env file
	utils.LoadEnv()

	tlsconfig := NewTLSConfig()

	// Initialize database connection
	if err := postgres.InitDb(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Websocket
	wsHub := websockets.StartWebsocketServer()

	// Get environment variables with defaults
	broker := utils.GetEnv("MQTT_BROKER", "mqtt://localhost:1883")
	clientID := utils.GetEnv("MQTT_CLIENT_ID", "home-security-backend")
	username := utils.GetEnv("MQTT_USERNAME", "")
	password := utils.GetEnv("MQTT_PASSWORD", "")

	log.Printf("Connecting to MQTT broker: %s\n", broker)
	log.Printf("Client ID: %s\n", clientID)

	// Configure MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID).SetTLSConfig(tlsconfig)

	if username != "" {
		opts.SetUsername(username)
	}
	if password != "" {
		opts.SetPassword(password)
	}

	opts.SetAutoReconnect(true)
	opts.SetDefaultPublishHandler(createMessageHandler(wsHub))
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	// Create and connect the client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}

	// Subscribe to sensor topics
	topic := "sensor/#"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	log.Printf("Subscribed to topic: %s\n", topic)

	// Keep the program running
	select {}
}

func NewTLSConfig() *tls.Config {
	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()
	pemCerts, err := os.ReadFile("../certs/CAfile.pem")
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}

	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair("../certs/client.crt", "../certs/client.key")
	if err != nil {
		panic(err)
	}

	// Just to print out the client certificate..
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		panic(err)
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: true,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}
}

func createMessageHandler(wsHub *websockets.WsHub) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		topic := msg.Topic()
		payload := msg.Payload()
		log.Printf("Received message: %s from topic: %s\n", payload, topic)

		if strings.HasPrefix(topic, "sensor/") && len(strings.Split(topic, "/")) >= 3 {
			parts := strings.Split(topic, "/")
			if len(parts) < 3 {
				log.Println("Invalid topic format: " + topic)
				return
			}
			sensorId := parts[1]

			var dat map[string]interface{}
			if err := json.Unmarshal(payload, &dat); err != nil {
				log.Printf("Error unmarshalling JSON: %v\n", err)
				return
			}
			// Get message timestamp
			messageTimestamp := dat["timestamp"].(float64)

			// Create a new sensor reading
			reading := &models.SensorReading{
				SensorID:         sensorId,
				Value:            0, // Assuming value is 0 for alarm messages
				Message:          string(payload),
				Timestamp:        time.Now(),
				MessageTimestamp: time.Unix(int64(messageTimestamp), 0),
			}

			if err := services.SensorReading.Create(reading); err != nil {
				log.Printf("Failed to create sensor reading: %v\n", err)
				return
			}

			wsHub.BroadcastMessage([]byte(reading.Message))

		}
	}
}
