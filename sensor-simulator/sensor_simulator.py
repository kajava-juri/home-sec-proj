import paho.mqtt.publish as mqtt_publish
import time
import json
# set up termial args
import argparse
parser = argparse.ArgumentParser(description='Sensor Simulator')

parser.add_argument('--host', type=str, default='localhost', help='MQTT broker host')
parser.add_argument('--port', type=int, default=1883, help='MQTT broker port')
parser.add_argument('--ca_cert', type=str, default=None, help='Path to CA certificate')
parser.add_argument('--cert', type=str, default=None, help='Path to client certificate')
parser.add_argument('--key', type=str, default=None, help='Path to client key')
parser.add_argument('--sensor_id', type=str, default='sensor_1', help='Sensor ID')

args = parser.parse_args()

sensor_id = args.sensor_id

mqtt_config = {
    "hostname": args.host,
    "port": args.port,
    "tls": {
        "ca_certs": args.ca_cert,
        "certfile": args.cert,
        "keyfile": args.key
    }
}

# helper function for mqtt
def publish_message(topic, payload):
    topic = f"sensor/{sensor_id}/{topic}"
    print(f"Publishing to {topic}: {payload}")
    try:
        mqtt_publish.single(topic, json.dumps(payload), **mqtt_config)
    except Exception as e:
        print(f"Failed to publish message: {e}")

def send_alarm():
    payload = {
        "timestamp": int(time.time()),
        "sensor_id": sensor_id,
        "message": "door opened",
        "severity": 3
    }
    publish_message("alarm", payload)

def send_status():
    payload = {
        "timestamp": int(time.time()),
        "sensor_id": sensor_id,
        "status": "active"
    }
    publish_message("status", payload)

if __name__ == "__main__":
    while True:
        send_alarm()
        time.sleep(5)  # wait for 5 seconds before sending the next alarm
        send_status()
        time.sleep(10)  # wait for 10 seconds before sending the next status update
        print("Sent alarm and status update.")