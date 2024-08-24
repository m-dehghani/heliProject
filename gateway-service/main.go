package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/m-dehghani/gateway-service/middleware"
	service "github.com/m-dehghani/gateway-service/service"
	"github.com/patrickmn/go-cache"
	"github.com/sony/gobreaker"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
)

// Circuit breaker settings
var cbSettings = gobreaker.Settings{
	Name:        "gRPC Circuit Breaker",
	MaxRequests: 5,
	Interval:    60 * time.Second,
	Timeout:     30 * time.Second,
}
var cb = gobreaker.NewCircuitBreaker(cbSettings)

// Debounce settings
var rateLimiter = rate.NewLimiter(rate.Every(100*time.Millisecond), 1)

var idempotencyCache = cache.New(5*time.Minute, 10*time.Minute)

// @title Gateway Service API
// @version 1.0
// @description This is the API documentation for the Gateway Service.
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	r := gin.Default()

	grpcClient := service.NewGRPCClient()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/register", func(c *gin.Context) {
		service.Register(c, grpcClient, cb)
	})

	r.POST("/login", func(c *gin.Context) {
		service.Login(c, grpcClient, cb)
	})

	r.POST("/logout", middleware.Authenticate, func(c *gin.Context) {
		service.Logout(c, grpcClient, cb)
	})

	r.POST("/deposit", middleware.Authenticate, func(c *gin.Context) {
		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		service.Deposit(c, grpcClient, cb)
	})

	r.POST("/withdraw", middleware.Authenticate, middleware.Idempotency, func(c *gin.Context) {
		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		service.Withdraw(c, grpcClient, cb)
	})

	r.GET("/balance", middleware.Authenticate, func(c *gin.Context) {
		service.Balance(c, grpcClient, cb)
	})

	r.GET("/transactions", middleware.Authenticate, func(c *gin.Context) {
		service.Transactions(c, grpcClient, cb)
	})

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "pong")
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
