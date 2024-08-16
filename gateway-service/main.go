package main

import (
	"context"
	pb "gateway-service/proto"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
)

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

	accountConn, err := grpc.NewClient("account-service:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer accountConn.Close()

	accountClient := pb.NewAccountServiceClient(accountConn)

	customerConn, err := grpc.NewClient("customer-service:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer customerConn.Close()

	customerClient := pb.NewCustomerServiceClient(customerConn)

	// Swagger endpoint

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register endpoint
	// @Summary Register a new user
	// @Description Register a new user with a username and password
	// @Tags User
	// @Accept json
	// @Produce json
	// @Param request body Credentials true "Register Request"
	// @Success 200 {object} gin.H
	// @Failure 400 {object} gin.H
	// @Failure 500 {object} gin.H
	// @Router /register [post]
	r.POST("/register", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		grpcReq := &pb.RegisterRequest{
			Username: req.Username,
			Password: req.Password,
		}

		grpcRes, err := customerClient.Register(context.Background(), grpcReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Call CreateAccount method in account-service
		accountReq := &pb.CreateAccountRequest{Customerid: grpcRes.Customerid}
		_, err = accountClient.CreateAccount(context.Background(), accountReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": grpcRes.Message, "customer-id": grpcRes.Customerid})
	})

	// Login endpoint
	// @Summary Login a user
	// @Description Login a user with a username and password
	// @Tags User
	// @Accept json
	// @Produce json
	// @Param request body Credentials true "Login Request"
	// @Success 200 {object} gin.H
	// @Failure 400 {object} gin.H
	// @Failure 401 {object} gin.H
	// @Failure 500 {object} gin.H
	// @Router /login [post]
	r.POST("/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		grpcReq := &pb.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		}

		grpcRes, err := customerClient.Login(context.Background(), grpcReq)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": grpcRes.Token, "message": grpcRes.Message})
	})

	// Logout endpoint
	// @Summary Logout a user
	// @Description Logout a user by invalidating the token
	// @Tags User
	// @Security BearerAuth
	// @Success 200 {object} gin.H
	// @Failure 401 {object} gin.H
	// @Failure 500 {object} gin.H
	// @Router /logout [post]
	r.POST("/logout", authenticate, func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		grpcReq := &pb.LogoutRequest{
			Token: token,
		}

		grpcRes, err := customerClient.Logout(context.Background(), grpcReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": grpcRes.Message})
	})

	// accountClient := pb.NewAccountServiceClient(conn)

	r.POST("/deposit", authenticate, func(c *gin.Context) {

		var req struct {
			CustomerID uint32  `json:"customer_id"`
			Amount     float64 `json:"amount"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		username, _ := c.Get("username")

		userValidationReq := &pb.VerifyCustomerIDRequest{
			Username:   username.(string),
			Customerid: req.CustomerID,
		}

		userValidationRes, err := customerClient.VerifyCustomerID(context.Background(), userValidationReq)
		if err != nil || !userValidationRes.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		grpcReq := &pb.DepositRequest{
			Customerid: req.CustomerID,
			Amount:     req.Amount,
		}

		grpcRes, err := accountClient.Deposit(context.Background(), grpcReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": grpcRes.Success,
			"message": grpcRes.Message,
		})
	})

	r.POST("/withdraw", authenticate, func(c *gin.Context) {

		var req struct {
			CustomerID uint32  `json:"customer_id"`
			Amount     float64 `json:"amount"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		username, _ := c.Get("username")

		userValidationReq := &pb.VerifyCustomerIDRequest{
			Username:   username.(string),
			Customerid: req.CustomerID,
		}

		userValidationRes, err := customerClient.VerifyCustomerID(context.Background(), userValidationReq)
		if err != nil || !userValidationRes.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		grpcReq := &pb.WithdrawRequest{
			Customerid: req.CustomerID,
			Amount:     req.Amount,
		}

		grpcRes, err := accountClient.Withdraw(context.Background(), grpcReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": grpcRes.Success,
			"message": grpcRes.Message,
		})
	})

	r.GET("/balance", authenticate, func(c *gin.Context) {

		customerID, err := strconv.ParseUint(c.Query("customer_id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		username, _ := c.Get("username")
		log.Println("in balance api customerID is ", customerID)

		userValidationReq := &pb.VerifyCustomerIDRequest{
			Username:   username.(string),
			Customerid: uint32(customerID),
		}
		log.Println("uservalidation request is: ", userValidationReq.Customerid, userValidationReq.Username)
		userValidationRes, err := customerClient.VerifyCustomerID(context.Background(), userValidationReq)
		if err != nil || !userValidationRes.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		grpcReq := &pb.BalanceInquiryRequest{
			Customerid: uint32(customerID),
		}

		grpcRes, err := accountClient.BalanceInquiry(context.Background(), grpcReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"balance": grpcRes.Balance,
			"message": grpcRes.Message,
		})
	})

	r.GET("/transactions", authenticate, func(c *gin.Context) {

		customerID, err := strconv.ParseUint(c.Query("customer_id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		username, _ := c.Get("username")

		userValidationReq := &pb.VerifyCustomerIDRequest{
			Username:   username.(string),
			Customerid: uint32(customerID),
		}

		userValidationRes, err := customerClient.VerifyCustomerID(context.Background(), userValidationReq)
		if err != nil || !userValidationRes.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		grpcReq := &pb.TransactionHistoryRequest{
			Customerid: uint32(customerID),
		}

		grpcRes, err := accountClient.TransactionHistory(context.Background(), grpcReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"transactions": grpcRes.Transactions,
			"message":      grpcRes.Message,
		})
	})

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "pong")
	})
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
