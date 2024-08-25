package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/m-dehghani/gateway-service/models/grpcclient"
	pb "github.com/m-dehghani/gateway-service/proto"
	"github.com/sony/gobreaker"
)

// BalanceRequest represents the request parameters for the Balance endpoint
type BalanceRequest struct {
	CustomerID uint32 `json:"customer_id"`
}

// TransactionsRequest represents the request parameters for the Transactions endpoint
type TransactionsRequest struct {
	CustomerID uint32 `json:"customer_id"`
}

// DepositRequest represents the request body for the Deposit endpoint
type DepositRequest struct {
	CustomerID uint32  `json:"customer_id"`
	Amount     float64 `json:"amount"`
}

// WithdrawRequest represents the request body for the Withdraw endpoint
type WithdrawRequest struct {
	CustomerID uint32  `json:"customer_id"`
	Amount     float64 `json:"amount"`
}

// @Summary Deposit money into account
// @Description Deposit a specified amount into the customer's account
// @Tags Account
// @Accept json
// @Produce json
// @Param Authorization header string true "Token"
// @Param request body DepositRequest true "Deposit Request"
// @Success 200
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /deposit [post]
func Deposit(c *gin.Context, grpcClient *grpcclient.GRPCClient, cb *gobreaker.CircuitBreaker) {
	var req DepositRequest
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
		return grpcClient.CustomerService.VerifyCustomerID(context.Background(), userValidationReq)
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
		return grpcClient.AccountService.Deposit(context.Background(), grpcReq)
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

// @Summary Withdraw money from account
// @Description Withdraw a specified amount from the customer's account
// @Tags Account
// @Accept json
// @Param Authorization header string true "Token"
// @Produce json
// @Param request body WithdrawRequest true "Withdraw Request"
// @Success 200
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /withdraw [post]
func Withdraw(c *gin.Context, grpcClient *grpcclient.GRPCClient, cb *gobreaker.CircuitBreaker) {
	var req WithdrawRequest
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
		return grpcClient.CustomerService.VerifyCustomerID(context.Background(), userValidationReq)
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
		return grpcClient.AccountService.Withdraw(context.Background(), grpcReq)
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

// @Summary Get account balance
// @Description Get the balance of the customer's account
// @Tags Account
// @Accept json
// @Produce json
// @Param Authorization header string true "Token"
// @Param customer_id query uint32 true "Customer ID"
// @Success 200
// @Failure 401
// @Failure 500
// @Router /balance [get]

func Balance(c *gin.Context, grpcClient *grpcclient.GRPCClient, cb *gobreaker.CircuitBreaker) {
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
		return grpcClient.CustomerService.VerifyCustomerID(context.Background(), userValidationReq)
	})
	if err != nil || !userValidationRes.(*pb.VerifyCustomerIDResponse).Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grpcReq := &pb.BalanceInquiryRequest{
		Customerid: uint32(customerID),
	}

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.AccountService.BalanceInquiry(context.Background(), grpcReq)
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

// @Summary Get transaction history
// @Param Authorization header string true "Token"
// @Description Get the transaction history of the customer's account
// @Tags Account
// @Accept json
// @Produce json
// @Param customer_id query uint32 true "Customer ID"
// @Success 200
// @Failure 401
// @Failure 500
// @Router /transactions [get]
func Transactions(c *gin.Context, grpcClient *grpcclient.GRPCClient, cb *gobreaker.CircuitBreaker) {
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
		return grpcClient.CustomerService.VerifyCustomerID(context.Background(), userValidationReq)
	})
	if err != nil || !userValidationRes.(*pb.VerifyCustomerIDResponse).Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	grpcReq := &pb.TransactionHistoryRequest{
		Customerid: uint32(customerID),
	}

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.AccountService.TransactionHistory(context.Background(), grpcReq)
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
