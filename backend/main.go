package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
    fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
    fmt.Printf("Connect lost: %v", err)
}

func main() {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        log.Println("Warning: Error loading .env file, using system environment variables")
    }

    tlsconfig := NewTLSConfig()

    // Get environment variables with defaults
    broker := getEnv("MQTT_BROKER", "mqtt://localhost:1883")
    clientID := getEnv("MQTT_CLIENT_ID", "home-security-backend")
    username := getEnv("MQTT_USERNAME", "")
    password := getEnv("MQTT_PASSWORD", "")

    fmt.Printf("Connecting to MQTT broker: %s\n", broker)
    fmt.Printf("Client ID: %s\n", clientID)


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
    
    opts.SetDefaultPublishHandler(messagePubHandler)
    opts.OnConnect = connectHandler
    opts.OnConnectionLost = connectLostHandler

    // Create and connect the client
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
    }

    // Subscribe to sensor topics
    topic := "sensor/+/alarm"
    token := client.Subscribe(topic, 1, nil)
    token.Wait()
    fmt.Printf("Subscribed to topic: %s\n", topic)

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
	fmt.Println(cert.Leaf)

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

// Helper function to get environment variable with a default value
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}