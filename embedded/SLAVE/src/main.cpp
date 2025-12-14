#include <Arduino.h>
#include <DHT.h>
#include <Wire.h>
#include <BH1750.h>
#include <SPI.h>
#include <LoRa.h>
#include <ArduinoJson.h>

#define DHTPIN 6
#define DHTTYPE DHT11
#define SOIL_PIN A0
#define RELAY_PIN 3  // This controls the pump relay

const char* SURVEY_POINT_ID = "ebf5507b-259c-4471-8be5-00f86b3cbaf0";

#define LORA_NSS 10
#define LORA_RST 9
#define LORA_DIO0 2

DHT dht(DHTPIN, DHTTYPE);
BH1750 lightMeter;

unsigned long lastSend = 0;

// ========== FORWARD DECLARATIONS ==========
void handleControlCommand(JsonDocument& doc);
void sendControlResponse(String device, String cmd, String status, String message);
void sendSensorData();

// ========== SETUP ==========
void setup() {
  Serial.begin(9600);
  
  dht.begin();
  Wire.begin();
  lightMeter.begin();
  
  pinMode(RELAY_PIN, OUTPUT);
  digitalWrite(RELAY_PIN, LOW);
  
  LoRa.setPins(LORA_NSS, LORA_RST, LORA_DIO0);
  
  Serial.print("Initializing LoRa...");
  if (!LoRa.begin(433E6)) {
    Serial.println("FAILED!");
    while (1);
  }
  
  LoRa.setSpreadingFactor(7);
  LoRa.setSignalBandwidth(125E3);
  LoRa.setCodingRate4(5);
  LoRa.setSyncWord(0x34);
  
  Serial.println("OK!");
  Serial.println("Survey Point: " + String(SURVEY_POINT_ID));
  Serial.println("Relay/Pump: Pin " + String(RELAY_PIN));
}

// ========== LOOP ==========
void loop() {
  // ===== RECEIVE CONTROL COMMANDS =====
  int packetSize = LoRa.parsePacket();
  if (packetSize) {
    String recv = "";
    while (LoRa.available()) {
      recv += (char)LoRa.read();
    }
    
    Serial.println("\n--- Received ---");
    Serial.println(recv);
    
    StaticJsonDocument<256> doc;
    if (deserializeJson(doc, recv) == DeserializationError::Ok) {
      String msgType = doc["type"] | "";
      String pointId = doc["survey_point_id"] | "";
      
      // Check if command is for this survey point
      if (msgType == "control" && pointId == SURVEY_POINT_ID) {
        handleControlCommand(doc);
      }
    }
  }
  
  // ===== SEND SENSOR DATA =====
  if (millis() - lastSend > 5000) {
    lastSend = millis();
    sendSensorData();
  }
}

// ========== CONTROL FUNCTIONS ==========
void handleControlCommand(JsonDocument& doc) {
  String device = doc["device"] | "";
  String cmd = doc["cmd"] | "";
  
  String status = "success";
  String message = "";
  
  // Handle both "relay" and "pump" as the same device
  if (device == "relay" || device == "pump") {
    if (cmd == "on") {
      digitalWrite(RELAY_PIN, HIGH);
      Serial.println("✓ Pump/Relay ON");
      message = device + " turned on";
    } 
    else if (cmd == "off") {
      digitalWrite(RELAY_PIN, LOW);
      Serial.println("✓ Pump/Relay OFF");
      message = device + " turned off";
    } 
    else {
      status = "failed";
      message = "Unknown command: " + cmd;
      Serial.println("✗ Unknown command");
    }
  } 
  else {
    status = "failed";
    message = "Unknown device: " + device;
    Serial.println("✗ Unknown device");
  }
  
  // Send response back to Gateway
  sendControlResponse(device, cmd, status, message);
}

void sendControlResponse(String device, String cmd, String status, String message) {
  StaticJsonDocument<384> doc;
  doc["type"] = "control_response";
  doc["survey_point_id"] = SURVEY_POINT_ID;
  doc["device"] = device;
  doc["cmd"] = cmd;
  doc["status"] = status;
  doc["message"] = message;
  
  String json;
  serializeJson(doc, json);
  
  LoRa.beginPacket();
  LoRa.print(json);
  LoRa.endPacket();
  
  Serial.println("Sent response: " + json);
}

// ========== SENSOR FUNCTIONS ==========
void sendSensorData() {
  float temp = dht.readTemperature();
  float hum = dht.readHumidity();
  float lux = lightMeter.readLightLevel();
  int soilRaw = analogRead(SOIL_PIN);
  int soil = map(soilRaw, 0, 1023, 0, 100);
  
  StaticJsonDocument<384> doc;
  doc["type"] = "sensor";
  doc["survey_point_id"] = SURVEY_POINT_ID;
  doc["temp"] = isnan(temp) ? 0 : temp;
  doc["hum"] = isnan(hum) ? 0 : hum;
  doc["lux"] = lux;
  doc["soil"] = soil;
  doc["relay"] = digitalRead(RELAY_PIN);  // Current pump/relay state
  
  String json;
  serializeJson(doc, json);
  
  LoRa.beginPacket();
  LoRa.print(json);
  LoRa.endPacket();
  
  Serial.println("Sent: " + json);
}
