package mqtt

import (
	"backend/database/models"
	"backend/database/services"
	"backend/pkg/websockets"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttClient struct {
	config MqttConfig
	client mqtt.Client
	wsHub  *websockets.WsHub
}

type MqttConfig struct {
	Broker   string
	ClientId string
	Username string
	Password string
	CAfile   string
}

func NewMqttClient(config MqttConfig, wsHub *websockets.WsHub) *MqttClient {
	return &MqttClient{
		config: config,
		wsHub:  wsHub,
	}
}

func (m *MqttClient) Connect() error {
	tlsconfig, err := m.NewTLSConfig(m.config.CAfile)
	if err != nil {
		return err
	}

	// Configure MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.config.Broker)
	opts.SetClientID(m.config.ClientId).SetTLSConfig(tlsconfig)

	if m.config.Username != "" {
		opts.SetUsername(m.config.Username)
	}
	if m.config.Password != "" {
		opts.SetPassword(m.config.Password)
	}

	opts.SetAutoReconnect(true)
	opts.SetDefaultPublishHandler(createMessageHandler(m.wsHub))
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	// Create and connect the client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.WaitTimeout(5 * time.Second) {
		if token.Error() != nil {
			log.Printf("Failed to connect to MQTT broker: %v", token.Error())
			log.Println("Application will continue running without MQTT connectivity")
			return token.Error()
		} else {
			log.Println("Connected to MQTT broker")
			// Subscribe to sensor topics
			topic := "sensor/#"
			token := client.Subscribe(topic, 1, nil)
			token.Wait()
			log.Printf("Subscribed to topic: %s\n", topic)
		}
	} else {
		log.Printf("Failed to connect to MQTT broker: timed out")
		log.Println("Application will continue running without MQTT connectivity")
		return token.Error()
	}

	return nil
}

func (m *MqttClient) NewTLSConfig(cafile string) (*tls.Config, error) {
	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()
	pemCerts, err := os.ReadFile(cafile)
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}

	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair("../certs/client.crt", "../certs/client.key")
	if err != nil {
		return nil, err
	}

	// Just to print out the client certificate..
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
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
	}, nil
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

			wsHub.BroadcastToTopic([]byte(reading.Message), "sensor/"+sensorId)
			wsHub.BroadcastToTopic([]byte(reading.Message), "sensors")

			if match, _ := regexp.MatchString(`sensor/\w*/alarm`, topic); match {
				wsHub.BroadcastToTopic(payload, "alerts")
			}

		}
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}
