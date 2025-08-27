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
	CAPath   string
	CertPath string
	KeyPath  string
}

func NewMqttClient(config MqttConfig, wsHub *websockets.WsHub) *MqttClient {
	return &MqttClient{
		config: config,
		wsHub:  wsHub,
	}
}

func (m *MqttClient) Connect() error {
	tlsconfig, err := m.NewTLSConfig(m.config.CAPath)
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
	m.client = mqtt.NewClient(opts)
	m.subscribeToTopics()

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
		parts := strings.Split(topic, "/")

		/*
			Alarms are device-specific:
				/sensor_hub/<device>/alarm
			Sensor events are specific to each sensor type + id:
				/sensor_hub/<device>/<type>/<sensor_id>/{sensor specific}
				for example: /sensor_hub/pico_w_1/door/front_door/{open|closed}
		*/
		if strings.HasPrefix(topic, "sensor_hub/") && len(parts) >= 3 {

			// The heartbeat can also be just a simple "1" or "0", not JSON
			// TODO: handle that case
			if (string(payload) == "1" || string(payload) == "0") && len(parts) == 3 && parts[2] == "heartbeat" {
				log.Println("Received simple heartbeat")
				return
			}
			var dat map[string]interface{}
			if err := json.Unmarshal(payload, &dat); err != nil {
				log.Printf("Error unmarshalling JSON: %v\n", err)
				return
			}

			deviceId := parts[1]
			foundDevice, err := services.Device.GetByID(deviceId)
			if err != nil {
				log.Printf("Error getting device: %v\n", err)
				return
			}

			if match, _ := regexp.MatchString(`sensor_hub/\w*/status`, topic); match {
				// Handle status messages
				foundDevice.AlarmStatus = dat["alarm_state"].(string)
				foundDevice.MqttStatus = dat["mqtt_status"].(string)
				foundDevice.UptimeMs = dat["uptime_ms"].(float64)
				foundDevice.WifiStatus = dat["wifi_status"].(string)
				if err := services.Device.Update(foundDevice); err != nil {
					log.Printf("Failed to update device status: %v\n", err)
					return
				}
			}

			// Handle alarm messages
			if match, _ := regexp.MatchString(`sensor_hub/\w*/alarm\b`, topic); match {
				event := ""
				if len(parts) >= 4 {
					event = parts[3]
				}
				alarm := &models.Alarm{
					DeviceID: foundDevice.ID,
					Payload:  string(payload),
					Event:    event,
				}
				if err := services.Alarm.Create(alarm); err != nil {
					log.Printf("Failed to create alarm: %v\n", err)
					return
				}
				wsHub.BroadcastToTopic([]byte(alarm.Payload), "alarms")

				//    [0]       [1]        [2]          [3]          [4]
				// sensor_hub/{device}/{sensor_type}/{sensor_id}/?{optional sensor specific event}?
			} else if match, _ := regexp.MatchString(`sensor_hub/\w*/\w*/\w*`, topic); match {

				// Handle sensor events
				sensorType := parts[2]
				sensorId := parts[3]
				sensorQuery := models.Sensor{
					Name:     sensorId,
					DeviceID: foundDevice.ID,
					Type:     sensorType,
				}
				sensor, err := services.Sensor.GetOrCreate(&sensorQuery)
				if err != nil {
					log.Printf("Error getting sensor: %v\n", err)
					return
				}

				sensorValue := ""
				if len(parts) > 4 {
					sensorValue = parts[4]
				}

				// Create a new sensor reading
				reading := &models.SensorReading{
					SensorID:  sensor.ID,
					DeviceID:  foundDevice.ID,
					Value:     sensorValue,
					Message:   string(payload),
					Topic:     topic,
					Timestamp: time.Now(),
					//MessageTimestamp: time.Unix(int64(messageTimestamp), 0),
				}

				if err := services.SensorReading.Create(reading); err != nil {
					log.Printf("Failed to create sensor reading: %v\n", err)
					return
				}

				wsHub.BroadcastToTopic([]byte(reading.Message), "sensor/"+sensorId)
				wsHub.BroadcastToTopic([]byte(reading.Message), "sensors")
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

func (m *MqttClient) subscribeToTopics() error {
	if token := m.client.Connect(); token.WaitTimeout(5 * time.Second) {
		if token.Error() != nil {
			log.Printf("Failed to connect to MQTT broker: %v", token.Error())
			log.Println("Application will continue running without MQTT connectivity")
			return token.Error()
		} else {
			log.Println("Connected to MQTT broker")
			// Subscribe to sensor topics
			topic := "sensor_hub/#"
			token := m.client.Subscribe(topic, 1, nil)
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
