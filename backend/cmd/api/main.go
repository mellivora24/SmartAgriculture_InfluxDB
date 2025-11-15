package main

import (
	"backend/infra/database"
	"backend/infra/network"
	"backend/internal/feature/farm"
	"backend/internal/feature/mcu"
	"backend/internal/feature/sensorData"
	"backend/internal/feature/surveyPoint"
	"backend/internal/feature/user"
	"backend/internal/shared"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Application struct {
	Config    *shared.Config
	Postgres  *database.PostgresDB
	InfluxDB  *database.InfluxDB
	MQTT      *network.MQTTClient
	WebSocket *network.WebSocketHub
	Router    *gin.Engine

	UserHandler        *user.Handler
	FarmHandler        *farm.Handler
	MCUHandler         *mcu.Handler
	SurveyPointHandler *surveyPoint.Handler
	SensorDataHandler  *sensorData.Handler
}

func main() {
	config := shared.LoadConfig()
	app, err := initializeApp(config)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	defer app.cleanup()

	app.setupRoutes()
	app.startServer()
}

func initializeApp(config *shared.Config) (*Application, error) {
	log.Println("Initializing application...")

	app := &Application{
		Config: config,
	}

	postgres, err := database.NewPostgresDB(&config.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	app.Postgres = postgres

	influxdb, err := database.NewInfluxDB(&config.InfluxDB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize InfluxDB: %w", err)
	}
	app.InfluxDB = influxdb

	mqttClient, err := network.NewMQTTClient(&config.MQTT)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MQTT: %w", err)
	}
	app.MQTT = mqttClient

	wsHub := network.NewWebSocketHub(&config.WebSocket)
	app.WebSocket = wsHub

	if err := app.initializeHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}

	if config.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	app.Router = gin.Default()

	log.Println("Application initialized successfully")
	return app, nil
}

func (app *Application) initializeHandlers() error {
	db := app.Postgres.GetDB()

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	app.UserHandler = user.NewHandler(userService)

	farmRepo := farm.NewRepository(db)
	farmService := farm.NewService(farmRepo)
	app.FarmHandler = farm.NewHandler(farmService)

	mcuRepo := mcu.NewRepository(db)
	mcuService := mcu.NewService(mcuRepo)
	app.MCUHandler = mcu.NewHandler(mcuService)

	surveyPointRepo := surveyPoint.NewRepository(db)
	surveyPointService := surveyPoint.NewService(surveyPointRepo)
	app.SurveyPointHandler = surveyPoint.NewHandler(surveyPointService)

	sensorDataRepo := sensorData.NewRepository(db, app.InfluxDB, app.Config.InfluxDB.Bucket)
	sensorDataService := sensorData.NewService(sensorDataRepo)
	app.SensorDataHandler = sensorData.NewHandler(sensorDataService)

	log.Println("All handlers initialized successfully")
	return nil
}

func (app *Application) setupRoutes() {
	app.Router.GET("/health", func(c *gin.Context) {
		health := map[string]interface{}{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		}

		if err := app.Postgres.HealthCheck(); err != nil {
			health["postgres"] = "error: " + err.Error()
			health["status"] = "degraded"
		} else {
			health["postgres"] = "ok"
		}

		if err := app.InfluxDB.HealthCheck(); err != nil {
			health["influxdb"] = "error: " + err.Error()
			health["status"] = "degraded"
		} else {
			health["influxdb"] = "ok"
		}

		if err := app.MQTT.HealthCheck(); err != nil {
			health["mqtt"] = "error: " + err.Error()
			health["status"] = "degraded"
		} else {
			health["mqtt"] = "ok"
		}

		health["websocket"] = map[string]interface{}{
			"connected_clients": app.WebSocket.GetClientCount(),
		}

		c.JSON(http.StatusOK, health)
	})

	app.Router.GET("/ws", func(c *gin.Context) {
		clientID := c.Query("client_id")
		if clientID == "" {
			clientID = fmt.Sprintf("client_%d", time.Now().UnixNano())
		}

		if err := app.WebSocket.HandleWebSocket(c, clientID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to upgrade connection",
			})
			return
		}
	})

	api := app.Router.Group("/api/v1")
	{
		app.UserHandler.RegisterRoutes(api)
		app.FarmHandler.RegisterRoutes(api)
		app.MCUHandler.RegisterRoutes(api)
		app.SurveyPointHandler.RegisterRoutes(api)
		app.SensorDataHandler.RegisterRoutes(api)

		api.GET("/stats/db", func(c *gin.Context) {
			stats := app.Postgres.GetStats()
			c.JSON(http.StatusOK, stats)
		})

		api.POST("/mqtt/publish", func(c *gin.Context) {
			var req struct {
				Topic   string      `json:"topic" binding:"required"`
				Message interface{} `json:"message" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := app.MQTT.Publish(req.Topic, req.Message, false); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "published"})
		})

		api.POST("/ws/broadcast", func(c *gin.Context) {
			var req struct {
				Message interface{} `json:"message" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := app.WebSocket.BroadcastJSON(req.Message); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "broadcasted"})
		})
	}

	log.Println("Routes configured")
}

func (app *Application) startServer() {
	addr := fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      app.Router,
		ReadTimeout:  app.Config.Server.ReadTimeout,
		WriteTimeout: app.Config.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func (app *Application) cleanup() {
	log.Println("Cleaning up resources...")

	if app.MQTT != nil {
		app.MQTT.Disconnect(250)
	}

	if app.InfluxDB != nil {
		app.InfluxDB.Close()
	}

	if app.Postgres != nil {
		if err := app.Postgres.Close(); err != nil {
			log.Printf("Error closing PostgreSQL: %v", err)
		}
	}

	log.Println("Cleanup completed")
}
