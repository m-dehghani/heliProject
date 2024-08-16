package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	pb "gateway-service/proto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockCustomerServiceClient struct {
	pb.CustomerServiceClient
}

type mockAccountServiceClient struct {
	pb.AccountServiceClient
}

func (m *mockCustomerServiceClient) Register(ctx context.Context, in *pb.RegisterRequest, opts ...grpc.CallOption) (*pb.RegisterResponse, error) {
	return &pb.RegisterResponse{Message: "registration successful"}, nil
}

func (m *mockCustomerServiceClient) Login(ctx context.Context, in *pb.LoginRequest, opts ...grpc.CallOption) (*pb.LoginResponse, error) {
	return &pb.LoginResponse{Token: "mock_token", Message: "login successful"}, nil
}

func (m *mockCustomerServiceClient) Logout(ctx context.Context, in *pb.LogoutRequest, opts ...grpc.CallOption) (*pb.LogoutResponse, error) {
	return &pb.LogoutResponse{Message: "logout successful"}, nil
}

func (m *mockAccountServiceClient) Deposit(ctx context.Context, in *pb.DepositRequest, opts ...grpc.CallOption) (*pb.DepositResponse, error) {
	return &pb.DepositResponse{Success: true, Message: "deposit successful"}, nil
}

func (m *mockAccountServiceClient) Withdraw(ctx context.Context, in *pb.WithdrawRequest, opts ...grpc.CallOption) (*pb.WithdrawResponse, error) {
	return &pb.WithdrawResponse{Success: true, Message: "withdraw successful"}, nil
}

func (m *mockAccountServiceClient) BalanceInquiry(ctx context.Context, in *pb.BalanceInquiryRequest, opts ...grpc.CallOption) (*pb.BalanceInquiryResponse, error) {
	return &pb.BalanceInquiryResponse{Balance: 100.0, Message: "balance inquiry successful"}, nil
}

func (m *mockAccountServiceClient) TransactionHistory(ctx context.Context, in *pb.TransactionHistoryRequest, opts ...grpc.CallOption) (*pb.TransactionHistoryResponse, error) {
	transactions := []*pb.Transaction{
		{Id: 1, Customerid: 1, Type: "deposit", Amount: 50.0, Date: "2023-01-01T00:00:00Z"},
		{Id: 2, Customerid: 1, Type: "withdraw", Amount: 30.0, Date: "2023-01-02T00:00:00Z"},
	}
	return &pb.TransactionHistoryResponse{Transactions: transactions, Message: "transaction history retrieved"}, nil
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	customerClient := &mockCustomerServiceClient{}
	accountClient := &mockAccountServiceClient{}

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

		c.JSON(http.StatusOK, gin.H{"message": grpcRes.Message})
	})

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

	r.POST("/logout", func(c *gin.Context) {
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

	r.POST("/deposit", func(c *gin.Context) {
		var req struct {
			CustomerID uint32  `json:"customer_id"`
			Amount     float64 `json:"amount"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	r.POST("/withdraw", func(c *gin.Context) {
		var req struct {
			CustomerID uint32  `json:"customer_id"`
			Amount     float64 `json:"amount"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	r.GET("/balance", func(c *gin.Context) {
		customerID, err := strconv.ParseUint(c.Query("customer_id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id is required"})
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

	r.GET("/transactions", func(c *gin.Context) {
		customerID, err := strconv.ParseUint(c.Query("customer_id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id is required"})
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

	return r
}

func TestRegister(t *testing.T) {
	router := setupRouter()

	reqBody := bytes.NewBufferString(`{"username":"testuser","password":"testpass"}`)
	req, _ := http.NewRequest("POST", "/register", reqBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "registration successful", response["message"])
}

func TestLogin(t *testing.T) {
	router := setupRouter()

	reqBody := bytes.NewBufferString(`{"username":"testuser","password":"testpass"}`)
	req, _ := http.NewRequest("POST", "/login", reqBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "mock_token", response["token"])
	assert.Equal(t, "login successful", response["message"])
}
