package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/m-dehghani/gateway-service/docs"
	"github.com/m-dehghani/gateway-service/middleware"
	"github.com/m-dehghani/gateway-service/models/grpcclient"
	"github.com/m-dehghani/gateway-service/models/handlers"
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

//	@title			HeliTech APIs
//	@version		1.0
//	@description	These apis are just for presentation
//	@termsOfService	http://terms.helitech.com

//	@support		Heli API Support
//	@contact.url	http://www.helitec.com
//	@contact.email	support@helitec.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

// @host	localhost:8080
func main() {
	r := gin.Default()

	grpcClient := grpcclient.NewGRPCClient()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/register", func(c *gin.Context) {
		handlers.Register(c, grpcClient, cb)
	})

	r.POST("/login", func(c *gin.Context) {
		handlers.Login(c, grpcClient, cb)
	})

	r.POST("/logout", middleware.Authenticate, func(c *gin.Context) {
		handlers.Logout(c, grpcClient, cb)
	})

	r.POST("/deposit", middleware.Authenticate, func(c *gin.Context) {
		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		handlers.Deposit(c, grpcClient, cb)
	})

	r.POST("/withdraw", middleware.Authenticate, middleware.Idempotency, func(c *gin.Context) {
		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		handlers.Withdraw(c, grpcClient, cb)
	})

	r.GET("/balance", middleware.Authenticate, func(c *gin.Context) {
		handlers.Balance(c, grpcClient, cb)
	})

	r.GET("/transactions", middleware.Authenticate, func(c *gin.Context) {
		handlers.Transactions(c, grpcClient, cb)
	})

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "pong")
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
