package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/personal-blog/backend/internal/config"
	"github.com/personal-blog/backend/internal/database"
	"github.com/personal-blog/backend/internal/handler"
)

func main() {
	// Load .env file (ignore error if file doesn't exist, e.g., in production)
	_ = godotenv.Load("../.env")

	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection pool
	ctx := context.Background()

	dbPool, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		if cfg.Environment == "production" {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		log.Printf("WARNING: PostgreSQL not available: %v", err)
	} else {
		log.Println("Connected to PostgreSQL successfully")
		defer dbPool.Close()
	}

	// Initialize Redis client
	redisClient, err := database.NewRedisClient(ctx, cfg)
	if err != nil {
		if cfg.Environment == "production" {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}
		log.Printf("WARNING: Redis not available: %v", err)
	} else {
		log.Println("Connected to Redis successfully")
		defer redisClient.Close()
	}

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint (no rate limit)
	healthHandler := handler.NewHealthHandler(dbPool, redisClient)
	api := router.Group("/api")
	healthHandler.RegisterRoutes(api)

	// Newsletter endpoints - DISABLED
	// Uncomment below to enable newsletter functionality
	/*
	if dbPool != nil {
		var emailService service.EmailService
		if cfg.SendGridAPIKey != "" {
			emailService = service.NewEmailService(
				cfg.SendGridAPIKey,
				cfg.EmailFrom,
				cfg.EmailFromName,
				cfg.SiteURL,
			)
			log.Println("Email service (SendGrid) initialized")
		} else {
			log.Println("WARNING: SendGrid API key not configured, emails will not be sent")
		}

		newsletterService := service.NewNewsletterService(dbPool, emailService)
		newsletterHandler := handler.NewNewsletterHandler(newsletterService)
		newsletterHandler.RegisterRoutes(api)
	}
	*/

	// API route groups (handlers will be registered in later tasks)
	// api.POST("/views/:slug", ...)
	// api.GET("/views/:slug", ...)
	// api.GET("/views", ...)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close database connections (deferred calls handle this, but log for clarity)
	log.Println("Closing database connections...")

	log.Println("Server exited gracefully")
}
