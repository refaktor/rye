#!/usr/bin/env python3

import paho.mqtt.client as mqtt
import threading
import time
from datetime import datetime

# Create a mutex (lock) for thread-safe file writing
file_mutex = threading.Lock()

# File path for saving messages
MESSAGES_FILE = "mqtt-messages.txt"

# Handler function that saves messages to file with mutex protection
def on_message(client, userdata, message):
    txt = message.payload.decode('utf-8')
    
    print("\n-NEW-MESSAGE----")
    print(txt)
    print("")
    print(f"Topic: {message.topic}")
    print(f"QoS: {message.qos}")
    print(f"Retained: {message.retain}")
    
    # Lock the mutex before writing to file
    with file_mutex:  # Python's 'with' statement automatically handles lock/unlock
        # Prepare message data for writing
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        topic = message.topic
        message_line = f"{timestamp} | Topic: {topic} | Message: {txt}\n"
        
        # Write to file (append mode)
        try:
            with open(MESSAGES_FILE, 'a', encoding='utf-8') as file:
                file.write(message_line)
            print(f"Message saved to file: {MESSAGES_FILE}")
        except Exception as e:
            print(f"Error writing to file: {e}")

def on_connect(client, userdata, flags, rc):
    if rc == 0:
        print("Connected to MQTT broker")
        # Subscribe to topic
        topic = "rye/test"
        client.subscribe(topic, qos=1)
        print(f"Subscribed to topic: {topic}")
    else:
        print(f"Failed to connect to MQTT broker, return code {rc}")

def on_disconnect(client, userdata, rc):
    print("Disconnected from MQTT broker")

def main():
    # MQTT client setup
    client = mqtt.Client()
    
    # Set callback functions
    client.on_connect = on_connect
    client.on_message = on_message
    client.on_disconnect = on_disconnect
    
    # Connect to MQTT broker
    try:
        client.connect("test.mosquitto.org", 1883, 60)
    except Exception as e:
        print(f"Couldn't connect to MQTT: {e}")
        return
    
    print(f"Listening for MQTT messages and saving to {MESSAGES_FILE} ...")
    print("Using mutex (threading.Lock) for thread-safe file writing")
    print("Press Ctrl+C to stop.")
    
    # Start the loop to process network traffic and dispatch callbacks
    client.loop_start()
    
    try:
        # Keep the main thread alive
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        print("\nShutting down...")
        client.loop_stop()
        client.disconnect()

if __name__ == "__main__":
    main()
