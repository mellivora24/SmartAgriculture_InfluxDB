package main

import (
	"backend/infra/database"
	"backend/infra/network"
	"backend/internal/feature/farm"
	"backend/internal/feature/mcu"
	"backend/internal/feature/sensorData"
	"backend/internal/feature/surveyPoint"
	"backend/internal/feature/threshold"
	"backend/internal/feature/user"
	mqttPkg "backend/internal/realtime/mqtt"
	realtimeShared "backend/internal/realtime/shared"
	wsPkg "backend/internal/realtime/websocket"
	"backend/internal/shared"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Application struct {
	Config    *shared.Config
	Postgres  *database.PostgresDB
	InfluxDB  *database.InfluxDB
	MQTT      *network.MQTTClient
	WebSocket *network.WebSocketHub

	Router *gin.Engine

	UserHandler        *user.Handler
	FarmHandler        *farm.Handler
	MCUHandler         *mcu.Handler
	SurveyPointHandler *surveyPoint.Handler
	SensorDataHandler  *sensorData.Handler
	ThresholdHandler   *threshold.Handler

	// Realtime handlers
	MQTTHandler *mqttPkg.Handler
	WSHandler   *wsPkg.Handler

	// Realtime services
	MQTTService realtimeShared.MQTTService
	WSService   realtimeShared.WebSocketService
}

func main() {
	config := shared.LoadConfig()
	app, err := initializeApp(config)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	defer app.cleanup()

	// Initialize realtime handlers
	if err := app.initializeRealtimeHandlers(); err != nil {
		log.Fatalf("Failed to initialize realtime handlers: %v", err)
	}

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

	app.Router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:5174"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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

	thresholdRepo := threshold.NewRepository(db)
	thresholdService := threshold.NewService(thresholdRepo)
	app.ThresholdHandler = threshold.NewHandler(thresholdService)

	// Initialize realtime services
	app.MQTTService = mqttPkg.NewService(app.MQTT)
	app.WSService = wsPkg.NewService(app.WebSocket)

	// Initialize realtime handlers
	app.MQTTHandler = mqttPkg.NewHandler(app.MQTTService, app.WSService, sensorDataService, thresholdService, surveyPointService)
	app.WSHandler = wsPkg.NewHandler(app.WSService, app.MQTTService, sensorDataService)

	log.Println("All handlers initialized successfully")
	return nil
}

func (app *Application) initializeRealtimeHandlers() error {
	log.Println("Initializing realtime handlers...")

	// Subscribe to MQTT topics
	if err := app.MQTTHandler.Init(); err != nil {
		return fmt.Errorf("failed to initialize MQTT handler: %w", err)
	}

	log.Println("Realtime handlers initialized successfully")
	return nil
}

func (app *Application) setupRoutes() {
	// Public routes
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
		app.WSHandler.HandleConnection(c.Writer, c.Request)
	})

	// API routes
	api := app.Router.Group("/api/v1")
	{
		// Public auth routes
		auth := api.Group("/users")
		{
			auth.POST("/register", app.UserHandler.Register)
			auth.POST("/login", app.UserHandler.Login)
		}

		protected := api.Group("")
		protected.Use(shared.AuthMiddleware())
		{
			// User routes (protected)
			users := protected.Group("/users")
			{
				users.GET("/:id", app.UserHandler.GetByID)
				users.PUT("/:id", app.UserHandler.Update)
				users.DELETE("/:id", app.UserHandler.Delete)
				users.GET("", app.UserHandler.List)
			}

			// Feature routes
			app.FarmHandler.RegisterRoutes(protected)
			app.MCUHandler.RegisterRoutes(protected)
			app.SurveyPointHandler.RegisterRoutes(protected)
			app.SensorDataHandler.RegisterRoutes(protected)
			app.ThresholdHandler.RegisterRoutes(protected)

			// Database stats
			protected.GET("/stats/db", func(c *gin.Context) {
				stats := app.Postgres.GetStats()
				c.JSON(http.StatusOK, stats)
			})

			// testing
			protected.POST("/mqtt/publish", func(c *gin.Context) {
				var req struct {
					Topic   string      `json:"topic" binding:"required"`
					Message interface{} `json:"message" binding:"required"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if err := app.MQTTService.PublishJSON(req.Topic, req.Message); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": "published"})
			})

			// testing
			protected.POST("/ws/broadcast", func(c *gin.Context) {
				var req struct {
					MCUCode string          `json:"mcu_code" binding:"required"`
					Topic   string          `json:"topic" binding:"required"`
					Message json.RawMessage `json:"message" binding:"required"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				wsMsg := realtimeShared.WSMessage{
					Topic:   req.Topic,
					Payload: req.Message,
				}

				if err := app.WSService.BroadcastToMCU(req.MCUCode, wsMsg); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": "broadcasted"})
			})
		}
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
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
