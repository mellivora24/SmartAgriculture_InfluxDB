package shared

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	InfluxDB  InfluxDBConfig
	MQTT      MQTTConfig
	WebSocket WebSocketConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Environment  string
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type InfluxDBConfig struct {
	URL           string
	Token         string
	Org           string
	Bucket        string
	Timeout       time.Duration
	BatchSize     int
	FlushInterval time.Duration
}

type MQTTConfig struct {
	Broker               string
	Port                 int
	ClientID             string
	Username             string
	Password             string
	QoS                  byte
	KeepAlive            int
	CleanSession         bool
	ConnectRetry         bool
	AutoReconnect        bool
	ConnectTimeout       time.Duration
	MaxReconnectInterval time.Duration
}

type WebSocketConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	MaxMessageSize  int64
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			Environment:  getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5435),
			User:            getEnv("DB_USER", "admin"),
			Password:        getEnv("DB_PASSWORD", "admin123456"),
			DBName:          getEnv("DB_NAME", "sa_db"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		InfluxDB: InfluxDBConfig{
			URL:           getEnv("INFLUX_URL", "http://localhost:8086"),
			Token:         getEnv("INFLUX_TOKEN", "ABCxyzQWErty123@456#789*"),
			Org:           getEnv("INFLUX_ORG", "smart_agriculture"),
			Bucket:        getEnv("INFLUX_BUCKET", "sa_realtime"),
			Timeout:       getEnvAsDuration("INFLUX_TIMEOUT", 20*time.Second),
			BatchSize:     getEnvAsInt("INFLUX_BATCH_SIZE", 100),
			FlushInterval: getEnvAsDuration("INFLUX_FLUSH_INTERVAL", 1*time.Second),
		},
		MQTT: MQTTConfig{
			Broker:               getEnv("MQTT_BROKER", "localhost"),
			Port:                 getEnvAsInt("MQTT_PORT", 1883),
			ClientID:             getEnv("MQTT_CLIENT_ID", "sa_backend_client"),
			Username:             getEnv("MQTT_USERNAME", "admin"),
			Password:             getEnv("MQTT_PASSWORD", "admin123456"),
			QoS:                  byte(getEnvAsInt("MQTT_QOS", 1)),
			KeepAlive:            getEnvAsInt("MQTT_KEEP_ALIVE", 60),
			CleanSession:         getEnvAsBool("MQTT_CLEAN_SESSION", true),
			ConnectRetry:         getEnvAsBool("MQTT_CONNECT_RETRY", true),
			AutoReconnect:        getEnvAsBool("MQTT_AUTO_RECONNECT", true),
			ConnectTimeout:       getEnvAsDuration("MQTT_CONNECT_TIMEOUT", 5*time.Second),
			MaxReconnectInterval: getEnvAsDuration("MQTT_MAX_RECONNECT_INTERVAL", 10*time.Minute),
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:  getEnvAsInt("WS_READ_BUFFER_SIZE", 1024),
			WriteBufferSize: getEnvAsInt("WS_WRITE_BUFFER_SIZE", 1024),
			ReadTimeout:     getEnvAsDuration("WS_READ_TIMEOUT", 60*time.Second),
			WriteTimeout:    getEnvAsDuration("WS_WRITE_TIMEOUT", 10*time.Second),
			PongWait:        getEnvAsDuration("WS_PONG_WAIT", 60*time.Second),
			PingPeriod:      getEnvAsDuration("WS_PING_PERIOD", 54*time.Second),
			MaxMessageSize:  int64(getEnvAsInt("WS_MAX_MESSAGE_SIZE", 512*1024)),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func (c *MQTTConfig) GetBrokerURL() string {
	return fmt.Sprintf("tcp://%s:%d", c.Broker, c.Port)
}
