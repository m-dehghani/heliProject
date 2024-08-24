package service

import (
	"context"
	"log"
	"net/http"
	"strconv"

	pb "github.com/m-dehghani/gateway-service/proto"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
)

type GRPCClient struct {
	accountClient  pb.AccountServiceClient
	customerClient pb.CustomerServiceClient
}

func NewGRPCClient() *GRPCClient {
	accountConn, err := grpc.Dial("account-service:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	customerConn, err := grpc.Dial("customer-service:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return &GRPCClient{
		accountClient:  pb.NewAccountServiceClient(accountConn),
		customerClient: pb.NewCustomerServiceClient(customerConn),
	}
}

func Register(c *gin.Context, grpcClient *GRPCClient, cb *gobreaker.CircuitBreaker) {
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

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.customerClient.Register(context.Background(), grpcReq)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	accountReq := &pb.CreateAccountRequest{Customerid: grpcRes.(*pb.RegisterResponse).Customerid}
	_, err = cb.Execute(func() (interface{}, error) {
		return grpcClient.accountClient.CreateAccount(context.Background(), accountReq)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": grpcRes.(*pb.RegisterResponse).Message, "customer-id": grpcRes.(*pb.RegisterResponse).Customerid})
}

func Login(c *gin.Context, grpcClient *GRPCClient, cb *gobreaker.CircuitBreaker) {
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

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.customerClient.Login(context.Background(), grpcReq)
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": grpcRes.(*pb.LoginResponse).Token, "message": grpcRes.(*pb.LoginResponse).Message})
}

func Logout(c *gin.Context, grpcClient *GRPCClient, cb *gobreaker.CircuitBreaker) {
	token := c.GetHeader("Authorization")

	grpcReq := &pb.LogoutRequest{
		Token: token,
	}

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.customerClient.Logout(context.Background(), grpcReq)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": grpcRes.(*pb.LogoutResponse).Message})
}

func Deposit(c *gin.Context, grpcClient *GRPCClient, cb *gobreaker.CircuitBreaker) {
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

	userValidationRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.customerClient.VerifyCustomerID(context.Background(), userValidationReq)
	})
	if err != nil || !userValidationRes.(*pb.VerifyCustomerIDResponse).Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grpcReq := &pb.DepositRequest{
		Customerid: req.CustomerID,
		Amount:     req.Amount,
	}

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.accountClient.Deposit(context.Background(), grpcReq)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": grpcRes.(*pb.DepositResponse).Success,
		"message": grpcRes.(*pb.DepositResponse).Message,
	})
}

func Withdraw(c *gin.Context, grpcClient *GRPCClient, cb *gobreaker.CircuitBreaker) {
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

	userValidationRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.customerClient.VerifyCustomerID(context.Background(), userValidationReq)
	})
	if err != nil || !userValidationRes.(*pb.VerifyCustomerIDResponse).Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grpcReq := &pb.WithdrawRequest{
		Customerid: req.CustomerID,
		Amount:     req.Amount,
	}

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.accountClient.Withdraw(context.Background(), grpcReq)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": grpcRes.(*pb.WithdrawResponse).Success,
		"message": grpcRes.(*pb.WithdrawResponse).Message,
	})
}

func Balance(c *gin.Context, grpcClient *GRPCClient, cb *gobreaker.CircuitBreaker) {
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

	userValidationRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.customerClient.VerifyCustomerID(context.Background(), userValidationReq)
	})
	if err != nil || !userValidationRes.(*pb.VerifyCustomerIDResponse).Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grpcReq := &pb.BalanceInquiryRequest{
		Customerid: uint32(customerID),
	}

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.accountClient.BalanceInquiry(context.Background(), grpcReq)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": grpcRes.(*pb.BalanceInquiryResponse).Balance,
		"message": grpcRes.(*pb.BalanceInquiryResponse).Message,
	})
}

func Transactions(c *gin.Context, grpcClient *GRPCClient, cb *gobreaker.CircuitBreaker) {
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

	userValidationRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.customerClient.VerifyCustomerID(context.Background(), userValidationReq)
	})
	if err != nil || !userValidationRes.(*pb.VerifyCustomerIDResponse).Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grpcReq := &pb.TransactionHistoryRequest{
		Customerid: uint32(customerID),
	}

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.accountClient.TransactionHistory(context.Background(), grpcReq)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": grpcRes.(*pb.TransactionHistoryResponse).Transactions,
		"message":      grpcRes.(*pb.TransactionHistoryResponse).Message,
	})
}
