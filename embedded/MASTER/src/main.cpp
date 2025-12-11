#include <Arduino.h>
#include <ESP8266WiFi.h>
#include <PubSubClient.h>
#include <SPI.h>
#include <LoRa.h>
#include <ArduinoJson.h>

// ========== CONFIG ==========
const char* WIFI_SSID = "TTL";
const char* WIFI_PASSWORD = "03022003";

const char* MQTT_BROKER = "172.20.10.4";
const int MQTT_PORT = 1883;
const char* MQTT_USER = "admin";
const char* MQTT_PASSWORD = "admin123456";

const char* USER_ID = "user123";
const char* MCU_CODE = "MCU001";

// ========== LORA CONFIG ==========
#define LORA_NSS D8
#define LORA_RST D0
#define LORA_DIO0 D1

WiFiClient espClient;
PubSubClient mqttClient(espClient);

// ========== TOPICS ==========
String topicSensorData;
String topicControlRequestUser;      // user/{USER_ID}/mcu/{MCU_CODE}/control/request
String topicControlRequestSystem;    // system/mcu/{MCU_CODE}/control/request
String topicControlResponse;
String topicAlert;

unsigned long lastReconnect = 0;
unsigned long lastHealthCheck = 0;

// ========== FORWARD DECLARATIONS ==========
void setupWiFi();
void setupLoRa();
void reconnectMQTT();
void mqttCallback(char* topic, byte* payload, unsigned int length);
void handleControlRequest(String message);
void receiveLoRaData();
void handleSensorData(JsonDocument& nodeDoc);
void handleControlResponse(JsonDocument& nodeDoc);
void publishControlResponse(String surveyPointId, String deviceName, String command, String status, String message);
void publishAlert(String title, String message, String severity);
void sendHealthCheck();

// ========== SETUP ==========
void setup() {
  Serial.begin(115200);
  Serial.println("\n\n=== ESP8266 LoRa-MQTT Gateway (Dual Topic) ===");
  
  // Build topics
  topicSensorData = "user/" + String(USER_ID) + "/mcu/" + String(MCU_CODE) + "/data";
  topicControlRequestUser = "user/" + String(USER_ID) + "/mcu/" + String(MCU_CODE) + "/control/request";
  topicControlRequestSystem = "system/mcu/" + String(MCU_CODE) + "/control/request";
  topicControlResponse = "user/" + String(USER_ID) + "/mcu/" + String(MCU_CODE) + "/control/response";
  topicAlert = "user/" + String(USER_ID) + "/mcu/" + String(MCU_CODE) + "/alert";
  
  setupWiFi();
  
  mqttClient.setServer(MQTT_BROKER, MQTT_PORT);
  mqttClient.setCallback(mqttCallback);
  mqttClient.setBufferSize(1024);
  
  setupLoRa();
  
  Serial.println("=== Ready ===\n");
}

// ========== LOOP ==========
void loop() {
  if (!mqttClient.connected()) {
    reconnectMQTT();
  }
  mqttClient.loop();
  
  if (millis() - lastHealthCheck > 30000) {
    lastHealthCheck = millis();
    sendHealthCheck();
  }
  
  receiveLoRaData();
  
  delay(10);
}

// ========== WIFI FUNCTIONS ==========
void setupWiFi() {
  Serial.print("Connecting to WiFi");
  WiFi.mode(WIFI_STA);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
  
  int attempts = 0;
  while (WiFi.status() != WL_CONNECTED && attempts < 30) {
    delay(500);
    Serial.print(".");
    attempts++;
  }
  
  if (WiFi.status() == WL_CONNECTED) {
    Serial.println("\n✓ WiFi connected");
    Serial.print("IP: ");
    Serial.println(WiFi.localIP());
  } else {
    Serial.println("\n✗ WiFi failed");
  }
}

// ========== LORA FUNCTIONS ==========
void setupLoRa() {
  LoRa.setPins(LORA_NSS, LORA_RST, LORA_DIO0);
  
  Serial.print("Initializing LoRa...");
  if (!LoRa.begin(433E6)) {
    Serial.println("FAILED!");
    return;
  }
  
  LoRa.setSpreadingFactor(7);
  LoRa.setSignalBandwidth(125E3);
  LoRa.setCodingRate4(5);
  LoRa.setSyncWord(0x34);
  
  Serial.println("OK!");
}

void receiveLoRaData() {
  int packetSize = LoRa.parsePacket();
  if (packetSize == 0) return;
  
  String recv = "";
  while (LoRa.available()) {
    recv += (char)LoRa.read();
  }
  
  int rssi = LoRa.packetRssi();
  float snr = LoRa.packetSnr();
  
  Serial.println("\n--- LoRa RX ---");
  Serial.printf("RSSI: %d dBm, SNR: %.2f dB\n", rssi, snr);
  Serial.println("Data: " + recv);
  
  StaticJsonDocument<512> nodeDoc;
  DeserializationError err = deserializeJson(nodeDoc, recv);
  
  if (err) {
    Serial.println("✗ Parse error: " + String(err.c_str()));
    return;
  }
  
  String msgType = nodeDoc["type"] | "sensor";
  
  if (msgType == "sensor") {
    handleSensorData(nodeDoc);
  } else if (msgType == "control_response") {
    handleControlResponse(nodeDoc);
  }
}

// ========== SENSOR DATA ==========
void handleSensorData(JsonDocument& nodeDoc) {
  String surveyPointId = nodeDoc["survey_point_id"] | "";
  
  if (surveyPointId.length() == 0) {
    Serial.println("✗ Missing survey_point_id");
    return;
  }
  
  StaticJsonDocument<768> mqttDoc;
  
  JsonObject payload = mqttDoc.createNestedObject("payload");
  payload["mcu_code"] = MCU_CODE;
  payload["survey_point_id"] = surveyPointId;
  
  if (nodeDoc.containsKey("temp")) {
    payload["temperature"] = nodeDoc["temp"].as<float>();
  }
  if (nodeDoc.containsKey("hum")) {
    payload["humidity"] = nodeDoc["hum"].as<float>();
  }
  if (nodeDoc.containsKey("soil")) {
    payload["soil_moisture"] = nodeDoc["soil"].as<float>();
  }
  if (nodeDoc.containsKey("lux")) {
    payload["light"] = nodeDoc["lux"].as<float>();
  }
  
  mqttDoc["topic"] = topicSensorData;
  mqttDoc["timestamp"] = millis();
  
  String json;
  serializeJson(mqttDoc, json);
  
  if (mqttClient.publish(topicSensorData.c_str(), json.c_str())) {
    Serial.println("✓ Published sensor data");
  } else {
    Serial.println("✗ Publish failed");
  }
}

// ========== CONTROL RESPONSE ==========
void handleControlResponse(JsonDocument& nodeDoc) {
  String surveyPointId = nodeDoc["survey_point_id"] | "";
  String deviceName = nodeDoc["device"] | "";
  String command = nodeDoc["cmd"] | "";
  String status = nodeDoc["status"] | "success";
  String message = nodeDoc["message"] | "";
  
  if (surveyPointId.length() == 0) {
    Serial.println("✗ Missing survey_point_id in response");
    return;
  }
  
  StaticJsonDocument<768> mqttDoc;
  
  JsonObject payload = mqttDoc.createNestedObject("payload");
  payload["survey_point_id"] = surveyPointId;
  payload["mcu_code"] = MCU_CODE;
  payload["device_name"] = deviceName;
  payload["command"] = command;
  payload["status"] = status;
  
  if (message.length() > 0) {
    payload["message"] = message;
  }
  
  payload["executed_at"] = millis();
  
  mqttDoc["topic"] = topicControlResponse;
  mqttDoc["timestamp"] = millis();
  
  String json;
  serializeJson(mqttDoc, json);
  
  if (mqttClient.publish(topicControlResponse.c_str(), json.c_str())) {
    Serial.println("✓ Published control response");
  } else {
    Serial.println("✗ Publish failed");
  }
}

// ========== MQTT FUNCTIONS ==========
void reconnectMQTT() {
  if (millis() - lastReconnect < 5000) return;
  lastReconnect = millis();
  
  Serial.print("Connecting to MQTT...");
  
  String clientId = "ESP8266_" + String(MCU_CODE) + "_" + String(random(0xffff), HEX);
  
  bool connected = false;
  if (strlen(MQTT_USER) > 0) {
    connected = mqttClient.connect(clientId.c_str(), MQTT_USER, MQTT_PASSWORD);
  } else {
    connected = mqttClient.connect(clientId.c_str());
  }
  
  if (connected) {
    Serial.println("✓ Connected!");
    
    // Subscribe to BOTH user and system control topics
    if (mqttClient.subscribe(topicControlRequestUser.c_str())) {
      Serial.println("✓ Subscribed: " + topicControlRequestUser);
    } else {
      Serial.println("✗ Subscribe failed (user topic)");
    }
    
    if (mqttClient.subscribe(topicControlRequestSystem.c_str())) {
      Serial.println("✓ Subscribed: " + topicControlRequestSystem);
    } else {
      Serial.println("✗ Subscribe failed (system topic)");
    }
    
    mqttClient.subscribe("/health/request");
    
  } else {
    Serial.print("✗ Failed, rc=");
    Serial.println(mqttClient.state());
  }
}

void mqttCallback(char* topic, byte* payload, unsigned int length) {
  String message = "";
  for (unsigned int i = 0; i < length; i++) {
    message += (char)payload[i];
  }
  
  Serial.println("\n--- MQTT RX ---");
  Serial.println("Topic: " + String(topic));
  Serial.println("Message: " + message);
  
  // Health check
  if (strcmp(topic, "/health/request") == 0) {
    mqttClient.publish("/health/response", "ok");
    Serial.println("✓ Health response sent");
    return;
  }
  
  // Control request from either user or system topic
  if (String(topic) == topicControlRequestUser || String(topic) == topicControlRequestSystem) {
    Serial.println("✓ Control request received from: " + String(topic));
    handleControlRequest(message);
  }
}

void handleControlRequest(String message) {
  StaticJsonDocument<768> doc;
  DeserializationError err = deserializeJson(doc, message);
  
  if (err) {
    Serial.println("✗ Parse error: " + String(err.c_str()));
    return;
  }
  
  // Extract nested payload
  JsonObject payload;
  if (doc.containsKey("payload")) {
    payload = doc["payload"].as<JsonObject>();
  } else {
    payload = doc.as<JsonObject>();
  }
  
  String surveyPointId = payload["survey_point_id"] | "";
  String deviceName = payload["device_name"] | "";
  String command = payload["command"] | "";
  
  if (surveyPointId.length() == 0 || deviceName.length() == 0 || command.length() == 0) {
    Serial.println("✗ Invalid control request");
    return;
  }
  
  Serial.printf("Control: %s -> %s (%s)\n", 
                deviceName.c_str(), command.c_str(), surveyPointId.c_str());
  
  // Build command for Node - map "pump" to "relay" if needed
  StaticJsonDocument<256> cmdDoc;
  cmdDoc["type"] = "control";
  cmdDoc["survey_point_id"] = surveyPointId;
  
  // Map device names: pump -> relay
  if (deviceName == "pump") {
    cmdDoc["device"] = "relay";
    Serial.println("ℹ Mapped 'pump' -> 'relay'");
  } else {
    cmdDoc["device"] = deviceName;
  }
  
  cmdDoc["cmd"] = command;
  
  // Add extra fields if present
  if (payload.containsKey("value")) {
    cmdDoc["value"] = payload["value"];
  }
  
  String cmdJson;
  serializeJson(cmdDoc, cmdJson);
  
  // Send via LoRa
  LoRa.beginPacket();
  LoRa.print(cmdJson);
  LoRa.endPacket();
  
  Serial.println("✓ Sent to Node: " + cmdJson);
  
  // Send immediate pending response to MQTT (use original device name)
  publishControlResponse(surveyPointId, deviceName, command, "pending", "Command sent to node");
}

void publishControlResponse(String surveyPointId, String deviceName, 
                           String command, String status, String message) {
  StaticJsonDocument<768> mqttDoc;
  
  JsonObject payload = mqttDoc.createNestedObject("payload");
  payload["survey_point_id"] = surveyPointId;
  payload["mcu_code"] = MCU_CODE;
  payload["device_name"] = deviceName;
  payload["command"] = command;
  payload["status"] = status;
  payload["message"] = message;
  payload["executed_at"] = millis();
  
  mqttDoc["topic"] = topicControlResponse;
  mqttDoc["timestamp"] = millis();
  
  String json;
  serializeJson(mqttDoc, json);
  
  if (mqttClient.publish(topicControlResponse.c_str(), json.c_str())) {
    Serial.println("✓ Published control response: " + status);
  }
}

void publishAlert(String title, String message, String severity) {
  StaticJsonDocument<768> mqttDoc;
  
  JsonObject payload = mqttDoc.createNestedObject("payload");
  payload["mcu_code"] = MCU_CODE;
  payload["title"] = title;
  payload["message"] = message;
  payload["severity"] = severity;
  payload["time"] = millis();
  
  mqttDoc["topic"] = topicAlert;
  mqttDoc["timestamp"] = millis();
  
  String json;
  serializeJson(mqttDoc, json);
  
  mqttClient.publish(topicAlert.c_str(), json.c_str());
}

void sendHealthCheck() {
  if (mqttClient.connected()) {
    Serial.println("♥ Health check");
  }
}