package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"net/http"

	"task-golang-db/handler"
	"task-golang-db/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Initialize database
	db := NewDatabase()
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get DB from GORM:", err)
	}
	defer sqlDB.Close()

	// Get signing key from .env
	signingKey := os.Getenv("SIGNING_KEY")

	// Initialize Gin router
	r := gin.Default()

	// Initialize Handlers
	authHandler := handler.NewAuth(db, []byte(signingKey))
	accountHandler := handler.NewAccount(db)
	transCatHandler := handler.NewTransactionCategory(db)
	transactionHandler := handler.NewTransaction(db)

	// Define Routes
	// Auth routes
	authRoute := r.Group("/auth")
	{
		authRoute.POST("/login", authHandler.Login)
		authRoute.POST("/upsert", authHandler.Upsert)
	}

	// Account routes
	accountRoutes := r.Group("/account")
	{
		accountRoutes.POST("/create", accountHandler.Create)
		accountRoutes.GET("/read/:id", accountHandler.Read)
		accountRoutes.PATCH("/update/:id", accountHandler.Update)
		accountRoutes.DELETE("/delete/:id", accountHandler.Delete)
		accountRoutes.GET("/list", accountHandler.List)
		accountRoutes.GET("/my", middleware.AuthMiddleware(signingKey), accountHandler.My)
		accountRoutes.POST("/topup", middleware.AuthMiddleware(signingKey), accountHandler.Topup)
		accountRoutes.GET("/balance", middleware.AuthMiddleware(signingKey), accountHandler.Balance)
		accountRoutes.POST("/transfer", middleware.AuthMiddleware(signingKey), accountHandler.Transfer)
		accountRoutes.GET("/mutation", middleware.AuthMiddleware(signingKey), accountHandler.Mutation)
	}

	// Transaction Category routes
	transCatRoutes := r.Group("/transaction-category")
	{
		transCatRoutes.POST("/create", middleware.AuthMiddleware(signingKey), transCatHandler.Create)
		transCatRoutes.GET("/read/:id", transCatHandler.Read)
		transCatRoutes.PATCH("/update/:id", transCatHandler.Update)
		transCatRoutes.DELETE("/delete/:id", transCatHandler.Delete)
		transCatRoutes.GET("/list", transCatHandler.List)
	}

	// Transaction routes
	transactionRoutes := r.Group("/transaction")
	{
		transactionRoutes.POST("/create", middleware.AuthMiddleware(signingKey), transactionHandler.NewTransaction)
		transactionRoutes.GET("/list", middleware.AuthMiddleware(signingKey), transactionHandler.TransactionList)
	}

	// Graceful shutdown setup
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

// NewDatabase initializes the database connection
func NewDatabase() *gorm.DB {
	dsn := os.Getenv("DATABASE")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get DB object: %v", err)
	}

	var currentDB string
	err = sqlDB.QueryRow("SELECT current_database()").Scan(&currentDB)
	if err != nil {
		log.Fatalf("Failed to query current database: %v", err)
	}

	log.Printf("Connected to database: %s\n", currentDB)

	return db
}
