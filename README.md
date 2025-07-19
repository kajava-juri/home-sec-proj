# Backend for Home Security Project
It will serve a REST API and handle MQTT messages that the sensors will send.
This repository also contains a simulator that sends fake data via MQTT.

## Running the simulator
1. Install dependencies:
```bash
go get github.com/joho/godotenv
go get github.com/eclipse/paho.mqtt.golang
```

2. Create a `.env` file in the `backend` directory with the following content:
```env
MQTT_BROKER=tcp://localhost:1883
MQTT_CLIENT_ID=home-security-backend
MQTT_USERNAME=
MQTT_PASSWORD=
```

3. Run the simulator:
```bash
python sensor_simulator.py --host raspberrypi.local --port 8883 --ca_cert ../certs/ca.crt --cert ../certs/client.crt --key ../certs/client.key
```

## Certificates

Documentation for setting up TLS certificates for Mosquitto MQTT broker.

## Creating TLS Certificates

```bash
# Create directory for certificates
mkdir ~/certs
cd ~/certs

### Create an extensions config file for x509 that will include the alternative names
``` bash
sudo nano extensions.cnf
```

copy the following content into the file (replace the alt_names section with your own values):

```ini
[ req ]
distinguished_name = dn          # empty stub, but must exist
req_extensions     = v3_req      # <-- add extensions to the CSR

[ dn ]                            # (can stay empty)

[ v3_req ]                        # picked up by the -extensions flag
basicConstraints = CA:FALSE
keyUsage         = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth            # what a TLS server needs
subjectAltName   = @alt_names            # pull actual names from the block below

[ alt_names ]
DNS.1 = raspberrypi.local
DNS.2 = pi.lan
DNS.3 = my-pi-alias
IP.1  = 192.168.1.42
IP.2  = 127.0.1.1
```

# 1. Create Certificate Authority (CA)
``` bash
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/C=EE/ST=Harjumaa/L=Tallinn/CN=MyCA/emailAddress=your-email@example.com"
```
# 2. Create server certificate
```bash
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr 
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 3650 -extensions v3_req -extfile <path to the above file>
```

# 3. Create client certificate
``` bash
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr -subj "/C=EE/ST=Harjumaa/L=Tallinn/CN=client1/emailAddress=your-email@example.com"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 3650
```

**Note:** Replace `your-email@example.com` with your actual email address. You can also modify the country (C), state (ST), and location (L) fields as needed.

## Step 3: Install Certificates for Mosquitto

```bash
# Copy certificates to mosquitto directory
sudo cp ~/certs/* /etc/mosquitto/certs/

# Set proper ownership and permissions
sudo chown -R mosquitto:mosquitto /etc/mosquitto/certs/
sudo chmod 644 /etc/mosquitto/certs/*.csr
sudo chmod 644 /etc/mosquitto/certs/*.crt
sudo chmod 600 /etc/mosquitto/certs/*.key
```

## Step 4: Configure Mosquitto

```bash
# Backup original config
sudo cp /etc/mosquitto/mosquitto.conf /etc/mosquitto/mosquitto.conf.backup

# Create new configuration
sudo tee /etc/mosquitto/mosquitto.conf << EOF
# Place your local configuration in /etc/mosquitto/conf.d/
#
# A full description of the configuration file is at
# /usr/share/doc/mosquitto/examples/mosquitto.conf.example

#per_listener_settings true

pid_file /run/mosquitto/mosquitto.pid

persistence true
persistence_location /var/lib/mosquitto/

log_dest file /var/log/mosquitto/mosquitto.log

include_dir /etc/mosquitto/conf.d

allow_anonymous true
listener 8883
#password_file /etc/mosquitto/passwd

cafile /etc/mosquitto/certs/ca.crt
certfile /etc/mosquitto/certs/server.crt
keyfile /etc/mosquitto/certs/server.key
require_certificate true
EOF
```

## Step 5: Restart and Test

```bash
# Restart Mosquitto service
sudo systemctl restart mosquitto

# Check service status
sudo systemctl status mosquitto

# Test TLS connection (subscriber)
mosquitto_sub -h localhost -p 8883 --cafile ~/certs/ca.crt --cert ~/certs/client.crt --key ~/certs/client.key -t test -d

# In another terminal, test publishing
mosquitto_pub -h localhost -p 8883 --cafile ~/certs/ca.crt --cert ~/certs/client.crt --key ~/certs/client.key -t test -m "Hello World"
```

## Troubleshooting

### Check Mosquitto logs:
```bash
sudo tail -f /var/log/mosquitto/mosquitto.log
```

### Common Issues:

**Permission denied errors:**
```bash
sudo chmod 644 /etc/mosquitto/certs/*.crt
sudo chmod 600 /etc/mosquitto/certs/*.key
sudo chown -R mosquitto:mosquitto /etc/mosquitto/certs/
```

**Certificate verification failed:**
- Ensure server certificate CN matches connection hostname (localhost)
- Verify certificates are signed by the same CA

**Connection refused - not authorized:**
- Check that `allow_anonymous true` is under the listener configuration
- Verify `per_listener_settings` configuration if used

### Verify certificates:
```bash
# Check certificate details
openssl x509 -in ~/certs/server.crt -text -noout | grep -A 1 "Subject:"
openssl x509 -in ~/certs/ca.crt -text -noout | grep -A 1 "Subject:"

# Verify certificate chain
openssl verify -CAfile ~/certs/ca.crt ~/certs/server.crt
openssl verify -CAfile ~/certs/ca.crt ~/certs/client.crt
```

## Security Notes

- Keep your private keys (*.key files) secure and never share them
- Consider using proper DNS names instead of localhost for production
- For production use, disable `allow_anonymous` and set up proper user authentication
- Regularly rotate certificates before they expire (10 years in this setup)
- Consider setting up a firewall to restrict access to port 8883

## File Locations

- **Certificates:** `/etc/mosquitto/certs/`
- **Configuration:** `/etc/mosquitto/mosquitto.conf`
- **Logs:** `/var/log/mosquitto/mosquitto.log`
- **Service:** `sudo systemctl status mosquitto`