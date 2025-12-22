#include <Arduino.h>
#include <DHT.h>
#include <Wire.h>
#include <BH1750.h>
#include <SPI.h>
#include <LoRa.h>

// ========== PIN CONFIG ==========
#define DHTPIN 6
#define DHTTYPE DHT11
#define SOIL_PIN A0
#define RELAY_PIN 3

// ========== LORA CONFIG ==========
#define LORA_NSS 10
#define LORA_RST 9
#define LORA_DIO0 2
#define LORA_FREQUENCY 433E6

// ========== NODE ID ==========
const char NODE_ID = 'A'; // Thay đổi: 'A', 'B', 'C' cho mỗi node

// ========== OBJECTS ==========
DHT dht(DHTPIN, DHTTYPE);
BH1750 lightMeter;

// ========== VARIABLES ==========
unsigned long lastSend = 0;
unsigned int packetCounter = 0;
bool waitingForAck = false;
unsigned long lastSendTime = 0;
String lastMessage = "";
int retryCount = 0;

// ========== SETTINGS ==========
const int SEND_INTERVAL = 5000;  // 5 giây
const int ACK_TIMEOUT = 2000;    // 2 giây
const int MAX_RETRY = 3;

void sendResponse(String status);
void sendSensorData();
void handleCommand(String device, String cmd);

void setup() {
  Serial.begin(9600);
  delay(1000);
  
  // Seed random
  randomSeed(analogRead(A1));
  
  Serial.println("===============================");
  Serial.println("  NANO SENDER - Simple Format");
  Serial.println("===============================");
  
  // Khởi tạo cảm biến
  Serial.print("DHT11...");
  dht.begin();
  Serial.println(" OK");
  
  Serial.print("BH1750...");
  Wire.begin();
  if (lightMeter.begin()) {
    Serial.println(" OK");
  } else {
    Serial.println(" LOI!");
  }
  
  pinMode(RELAY_PIN, OUTPUT);
  digitalWrite(RELAY_PIN, LOW);
  Serial.println("Relay... OK");
  
  // Khởi tạo LoRa
  Serial.print("LoRa...");
  LoRa.setPins(LORA_NSS, LORA_RST, LORA_DIO0);
  
  if (!LoRa.begin(LORA_FREQUENCY)) {
    Serial.println(" THAT BAI!");
    while (1) {
      delay(1000);
      Serial.println("Kiem tra ket noi LoRa!");
    }
  }
  
  LoRa.setSpreadingFactor(7);
  LoRa.setSignalBandwidth(125E3);
  LoRa.setCodingRate4(5);
  LoRa.setSyncWord(0x12);
  LoRa.setTxPower(17);
  
  Serial.println(" OK");
  Serial.println("===============================");
  Serial.println("Node: " + String(NODE_ID));
  Serial.println("===============================\n");
}

void loop() {
  // Nhận ACK hoặc lệnh điều khiển
  int packetSize = LoRa.parsePacket();
  if (packetSize) {
    String recv = "";
    while (LoRa.available()) {
      recv += (char)LoRa.read();
    }
    
    Serial.println("[RECV] " + recv);
    
    // Kiểm tra ACK: ACK,A,123
    if (recv.startsWith("ACK,")) {
      int idx1 = recv.indexOf(',');
      int idx2 = recv.indexOf(',', idx1 + 1);
      
      char nodeId = recv.charAt(idx1 + 1);
      int ackPacketId = recv.substring(idx2 + 1).toInt();
      
      if (nodeId == NODE_ID && ackPacketId == (packetCounter - 1)) {
        waitingForAck = false;
        Serial.println("[ACK] Thanh cong!");
      }
    }
    // Kiểm tra lệnh: CMD,A,relay,on
    else if (recv.startsWith("CMD,")) {
      int idx1 = recv.indexOf(',');
      int idx2 = recv.indexOf(',', idx1 + 1);
      int idx3 = recv.indexOf(',', idx2 + 1);
      
      char nodeId = recv.charAt(idx1 + 1);
      String device = recv.substring(idx2 + 1, idx3);
      String cmd = recv.substring(idx3 + 1);
      
      if (nodeId == NODE_ID) {
        handleCommand(device, cmd);
      }
    }
  }
  
  // Kiểm tra timeout và retry
  if (waitingForAck && (millis() - lastSendTime > ACK_TIMEOUT)) {
    if (retryCount < MAX_RETRY) {
      retryCount++;
      Serial.println("[RETRY] " + String(retryCount) + "/" + String(MAX_RETRY));
      
      // Random delay để tránh collision
      delay(random(50, 300));
      
      LoRa.beginPacket();
      LoRa.print(lastMessage);
      LoRa.endPacket(true);
      
      lastSendTime = millis();
    } else {
      Serial.println("[ERROR] Khong nhan ACK!");
      waitingForAck = false;
      retryCount = 0;
    }
  }
  
  // Gửi dữ liệu định kỳ
  if (!waitingForAck && (millis() - lastSend >= SEND_INTERVAL)) {
    sendSensorData();
  }
}

void sendSensorData() {
  // Đọc cảm biến
  float temp = dht.readTemperature();
  float hum = dht.readHumidity();
  float lux = lightMeter.readLightLevel();
  int soilRaw = analogRead(SOIL_PIN);
  int soil = map(soilRaw, 0, 1023, 0, 100);
  int relay = digitalRead(RELAY_PIN);
  
  // Xử lý giá trị NaN
  if (isnan(temp)) temp = 0;
  if (isnan(hum)) hum = 0;
  if (lux < 0) lux = 0;
  
  // Format: NodeID,PacketID,Temp,Hum,Lux,Soil,Relay
  // Ví dụ: A,123,25.5,60.2,450.0,65,1
  String message = String(NODE_ID) + "," +
                   String(packetCounter) + "," +
                   String(temp, 1) + "," +
                   String(hum, 1) + "," +
                   String(lux, 1) + "," +
                   String(soil) + "," +
                   String(relay);
  
  Serial.println("\n[SEND #" + String(packetCounter) + "]");
  Serial.println("  T=" + String(temp, 1) + "C  H=" + String(hum, 1) + "%");
  Serial.println("  L=" + String(lux, 1) + "lx  S=" + String(soil) + "%  R=" + String(relay));
  Serial.println("  Data: " + message);
  
  // Random delay nhỏ để tránh collision
  delay(random(10, 100));
  
  // Gửi
  LoRa.beginPacket();
  LoRa.print(message);
  LoRa.endPacket(true);
  
  lastMessage = message;
  lastSendTime = millis();
  lastSend = millis();
  waitingForAck = true;
  retryCount = 0;
  packetCounter++;
}

void handleCommand(String device, String cmd) {
  Serial.println("[CMD] " + device + " -> " + cmd);
  
  if (device == "relay" || device == "pump") {
    if (cmd == "on") {
      digitalWrite(RELAY_PIN, HIGH);
      Serial.println("  => Bat relay");
      sendResponse("OK");
    } 
    else if (cmd == "off") {
      digitalWrite(RELAY_PIN, LOW);
      Serial.println("  => Tat relay");
      sendResponse("OK");
    }
    else {
      sendResponse("ERROR");
    }
  }
}

void sendResponse(String status) {
  // Format: RESP,NodeID,Status
  // Ví dụ: RESP,A,OK
  String response = "RESP," + String(NODE_ID) + "," + status;
  
  delay(random(10, 50));
  
  LoRa.beginPacket();
  LoRa.print(response);
  LoRa.endPacket(true);
  
  Serial.println("[RESP] " + response);
}
