#include <Arduino.h>
#include <ESP8266WiFi.h>
#include <PubSubClient.h>
#include <SPI.h>
#include <LoRa.h>
#include <ArduinoJson.h>

#define WIFI_SSID "WiFi"
#define WIFI_PASSWORD "123456789"

#define MQTT_SERVER "192.168.137.248"
#define MQTT_PORT 1883
#define MQTT_USER "admin"
#define MQTT_PASSWORD "admin123456"

#define USER_ID "c4e56f18-ec8d-4a09-8555-fd88c7886c6e"
#define MCU_CODE "439112"

struct NodeConfig {
  char nodeId;
  const char* surveyPointId;
};

NodeConfig nodeMapping[] = {
  {'A', "7d3c6820-1699-42b7-a6d2-58ece4957c38"},
  {'B', "65c88cda-07c5-4aca-bdf6-9defadf0dc26"},
  {'C', "c4e56f18-ec8d-4a09-8555-fd88c7886c6e"}
};

const int NODE_COUNT = sizeof(nodeMapping) / sizeof(NodeConfig);

#define LORA_NSS D8
#define LORA_RST D0
#define LORA_DIO0 D1
#define LORA_FREQUENCY 433E6

WiFiClient espClient;
PubSubClient mqttClient(espClient);

unsigned long lastReconnect = 0;
unsigned long totalReceived = 0;

struct PendingCommand {
  char nodeId;
  String surveyPointId;
  String deviceName;
  String command;
  int expectedState;
  bool pending;
  unsigned long timestamp;
};

PendingCommand pendingCommands[10];
int pendingCount = 0;

void setupWiFi();
void setupMQTT();
void reconnectMQTT();
void mqttCallback(char* topic, byte* payload, unsigned int length);
void parseSensorData(String data, int rssi, float snr);
void sendToMQTT(char nodeId, int packetId, float temp, float hum, float lux, int soil, int relay, int rssi, float snr);
void handleControlRequest(JsonDocument& doc);
void sendControlToNode(char nodeId, String device, String command);
void checkPendingCommands(char nodeId, int relayState);
String getNodeId(const char* surveyPointId);
const char* getSurveyPointId(char nodeId);

void setup() {
  Serial.begin(115200);
  delay(1000);

  setupWiFi();
  setupMQTT();

  LoRa.setPins(LORA_NSS, LORA_RST, LORA_DIO0);

  if (!LoRa.begin(LORA_FREQUENCY)) {
    Serial.println(" THAT BAI!");
    while (1) delay(1000);
  }

  LoRa.setSpreadingFactor(7);
  LoRa.setSignalBandwidth(125E3);
  LoRa.setCodingRate4(5);
  LoRa.setSyncWord(0x12);
  LoRa.setTxPower(20);

  Serial.println("===============================");
  Serial.println("User ID: " + String(USER_ID));
  Serial.println("MCU Code: " + String(MCU_CODE));
  Serial.println("===============================");
}

void loop() {
  if (!mqttClient.connected()) {
    if (millis() - lastReconnect > 5000) {
      reconnectMQTT();
      lastReconnect = millis();
    }
  }
  mqttClient.loop();

  int packetSize = LoRa.parsePacket();
  if (packetSize) {
    String data = "";
    while (LoRa.available()) {
      data += (char)LoRa.read();
    }

    int rssi = LoRa.packetRssi();
    float snr = LoRa.packetSnr();

    Serial.println("[LoRa RECV] " + data);

    if (data.startsWith("RESP,")) {
      int idx1 = data.indexOf(',');
      int idx2 = data.indexOf(',', idx1 + 1);

      if (idx1 > 0 && idx2 > 0) {
        char nodeId = data.charAt(idx1 + 1);
        String status = data.substring(idx2 + 1);

        Serial.println("[RESP] NodeID=" + String(nodeId) + " Status=" + status);

        for (int i = 0; i < pendingCount; i++) {
          if (pendingCommands[i].pending && pendingCommands[i].nodeId == nodeId) {
            Serial.println("[RESP] Found pending command for node " + String(nodeId));

            StaticJsonDocument<512> doc;
            doc["topic"] = "control_response";

            JsonObject payload = doc.createNestedObject("payload");
            payload["survey_point_id"] = pendingCommands[i].surveyPointId;
            payload["mcu_code"] = MCU_CODE;
            payload["device_name"] = pendingCommands[i].deviceName;
            payload["command"] = pendingCommands[i].command;
            payload["status"] = (status == "OK") ? "success" : "failed";

            String message;
            serializeJson(doc, message);

            String topic = "user/" + String(USER_ID) + "/mcu/" + String(MCU_CODE) + "/control/response";
            
            Serial.println("[MQTT] Publishing to: " + topic);
            Serial.println("[MQTT] Message: " + message);
            
            // IMPORTANT: Check MQTT connection before publishing
            if (!mqttClient.connected()) {
              Serial.println("[MQTT] Not connected! Reconnecting...");
              reconnectMQTT();
              delay(200); // Wait for reconnect
            }
            
            if (mqttClient.connected()) {
              bool published = mqttClient.publish(topic.c_str(), message.c_str());
              if (published) {
                Serial.println("[MQTT] Published successfully");
                pendingCommands[i].pending = false;
              } else {
                Serial.println("[MQTT] Publish FAILED!");
                // Keep pending to retry later
              }
            } else {
              Serial.println("[MQTT] Still not connected after reconnect!");
              // Keep pending to retry later
            }

            break;
          }
        }
      }
    } else if (data.indexOf(',') > 0) {
      parseSensorData(data, rssi, snr);
    }
  }

  for (int i = 0; i < pendingCount; i++) {
    if (pendingCommands[i].pending &&
        millis() - pendingCommands[i].timestamp > 30000) {
      pendingCommands[i].pending = false;
    }
  }
}

void setupWiFi() {
  WiFi.mode(WIFI_STA);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

  int attempts = 0;
  while (WiFi.status() != WL_CONNECTED && attempts < 20) {
    delay(500);
    attempts++;
  }
}

void setupMQTT() {
  mqttClient.setServer(MQTT_SERVER, MQTT_PORT);
  mqttClient.setCallback(mqttCallback);
  mqttClient.setBufferSize(1024);
}

void reconnectMQTT() {
  if (WiFi.status() != WL_CONNECTED) return;

  String clientId = "ESP8266_" + String(MCU_CODE);
  bool connected = mqttClient.connect(clientId.c_str(), MQTT_USER, MQTT_PASSWORD);

  if (connected) {
    Serial.println("[MQTT] Connected successfully");
    
    String topic1 = "system/mcu/" + String(MCU_CODE) + "/control/request";
    String topic2 = "user/" + String(USER_ID) + "/mcu/" + String(MCU_CODE) + "/control/request";
    
    mqttClient.subscribe(topic1.c_str());
    Serial.println("[MQTT] Subscribed to: " + topic1);
    
    mqttClient.subscribe(topic2.c_str());
    Serial.println("[MQTT] Subscribed to: " + topic2);
  } else {
    Serial.println("[MQTT] Connection failed, rc=" + String(mqttClient.state()));
  }
}

void mqttCallback(char* topic, byte* payload, unsigned int length) {
  String message = "";
  for (unsigned int i = 0; i < length; i++) message += (char)payload[i];

  Serial.println("\n[MQTT] Message received:");
  Serial.println("  Topic: " + String(topic));
  Serial.println("  Payload: " + message);

  StaticJsonDocument<512> doc;
  DeserializationError error = deserializeJson(doc, message);
  if (error) {
    Serial.println("[MQTT] JSON parse error: " + String(error.c_str()));
    return;
  }

  handleControlRequest(doc);
}

void handleControlRequest(JsonDocument& doc) {
  JsonObject payload = doc.containsKey("payload") ? doc["payload"] : doc.as<JsonObject>();

  const char* surveyPointId = payload["survey_point_id"];
  const char* deviceName = payload["device_name"];
  const char* command = payload["command"];

  Serial.println("[Control Request] Received:");
  Serial.println("  SurveyPointID: " + String(surveyPointId ? surveyPointId : "NULL"));
  Serial.println("  DeviceName: " + String(deviceName ? deviceName : "NULL"));
  Serial.println("  Command: " + String(command ? command : "NULL"));

  if (!surveyPointId || !deviceName || !command) {
    Serial.println("[Control Request] Missing parameters!");
    return;
  }

  String nodeIdStr = getNodeId(surveyPointId);
  if (nodeIdStr.length() == 0) {
    Serial.println("[Control Request] Unknown surveyPointId: " + String(surveyPointId));
    return;
  }

  char nodeId = nodeIdStr.charAt(0);
  Serial.println("[Control Request] Target NodeID: " + String(nodeId));
  
  sendControlToNode(nodeId, deviceName, command);

  if (pendingCount < 10) {
    pendingCommands[pendingCount++] = {
      nodeId,
      surveyPointId,
      deviceName,
      command,
      String(command) == "on",
      true,
      millis()
    };
    Serial.println("[Control Request] Added to pending list (count=" + String(pendingCount) + ")");
  } else {
    Serial.println("[Control Request] WARNING: Pending list full!");
  }
}

void sendControlToNode(char nodeId, String device, String command) {
  String message = "CMD," + String(nodeId) + "," + device + "," + command;
  
  Serial.println("[LoRa SEND] " + message);
  
  LoRa.beginPacket();
  LoRa.print(message);
  LoRa.endPacket();
  
  Serial.println("[LoRa SEND] Command sent successfully");
}

void parseSensorData(String data, int rssi, float snr) {
  int comma[6], count = 0;
  for (int i = 0; i < data.length() && count < 6; i++)
    if (data[i] == ',') comma[count++] = i;
  if (count != 6) return;

  char nodeId = data[0];
  int packetId = data.substring(comma[0] + 1, comma[1]).toInt();
  float temp = data.substring(comma[1] + 1, comma[2]).toFloat();
  float hum = data.substring(comma[2] + 1, comma[3]).toFloat();
  float lux = data.substring(comma[3] + 1, comma[4]).toFloat();
  int soil = data.substring(comma[4] + 1, comma[5]).toInt();
  int relay = data.substring(comma[5] + 1).toInt();

  LoRa.beginPacket();
  LoRa.print("ACK," + String(nodeId) + "," + String(packetId));
  LoRa.endPacket();

  sendToMQTT(nodeId, packetId, temp, hum, lux, soil, relay, rssi, snr);
  checkPendingCommands(nodeId, relay);
}

void sendToMQTT(char nodeId, int packetId, float temp, float hum, float lux, int soil, int relay, int rssi, float snr) {
  if (!mqttClient.connected()) return;

  const char* surveyPointId = getSurveyPointId(nodeId);
  if (!surveyPointId) return;

  StaticJsonDocument<768> doc;
  doc["topic"] = "sensor_data";

  JsonObject payload = doc.createNestedObject("payload");
  payload["mcu_code"] = MCU_CODE;
  payload["survey_point_id"] = surveyPointId;
  payload["temperature"] = temp;
  payload["humidity"] = hum;
  payload["soil_moisture"] = soil;
  payload["light"] = lux;

  JsonObject extra = payload.createNestedObject("extra");
  extra["packet_id"] = packetId;
  extra["relay_state"] = relay;
  extra["signal_strength"] = rssi;
  extra["snr"] = snr;

  String message;
  serializeJson(doc, message);

  mqttClient.publish(("user/" + String(USER_ID) + "/mcu/" + String(MCU_CODE) + "/data").c_str(), message.c_str());
}

void checkPendingCommands(char nodeId, int relayState) {
  for (int i = 0; i < pendingCount; i++) {
    if (pendingCommands[i].pending &&
        pendingCommands[i].nodeId == nodeId &&
        pendingCommands[i].expectedState == relayState) {

      StaticJsonDocument<512> doc;
      doc["topic"] = "control_response";

      JsonObject payload = doc.createNestedObject("payload");
      payload["survey_point_id"] = pendingCommands[i].surveyPointId;
      payload["mcu_code"] = MCU_CODE;
      payload["device_name"] = pendingCommands[i].deviceName;
      payload["command"] = pendingCommands[i].command;
      payload["status"] = "success";

      String message;
      serializeJson(doc, message);

      mqttClient.publish(
        ("user/" + String(USER_ID) + "/mcu/" + String(MCU_CODE) + "/control/response").c_str(),
        message.c_str()
      );

      pendingCommands[i].pending = false;
    }
  }
}

String getNodeId(const char* surveyPointId) {
  for (int i = 0; i < NODE_COUNT; i++)
    if (strcmp(nodeMapping[i].surveyPointId, surveyPointId) == 0)
      return String(nodeMapping[i].nodeId);
  return "";
}

const char* getSurveyPointId(char nodeId) {
  for (int i = 0; i < NODE_COUNT; i++)
    if (nodeMapping[i].nodeId == nodeId)
      return nodeMapping[i].surveyPointId;
  return nullptr;
}
