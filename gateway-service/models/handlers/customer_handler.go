package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/m-dehghani/gateway-service/models/grpcclient"
	pb "github.com/m-dehghani/gateway-service/proto"
	"github.com/sony/gobreaker"
)

// RegisterRequest represents the request body for the Register endpoint
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest represents the request body for the Login endpoint
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// @Summary Register a new user
// @Description Register a new user with username and password
// @Tags Customer
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register Request"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /register [post]
func Register(c *gin.Context, grpcClient *grpcclient.GRPCClient, cb *gobreaker.CircuitBreaker) {
	var req RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &pb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	}

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.CustomerService.Register(context.Background(), grpcReq)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	accountReq := &pb.CreateAccountRequest{Customerid: grpcRes.(*pb.RegisterResponse).Customerid}
	_, err = cb.Execute(func() (interface{}, error) {
		return grpcClient.AccountService.CreateAccount(context.Background(), accountReq)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": grpcRes.(*pb.RegisterResponse).Message, "customer-id": grpcRes.(*pb.RegisterResponse).Customerid})
}

// @Summary Login a user
// @Description Login a user with username and password
// @Tags Customer
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Request"
// @Success 200
// @Failure 400
// @Failure 401
// @Router /login [post]
func Login(c *gin.Context, grpcClient *grpcclient.GRPCClient, cb *gobreaker.CircuitBreaker) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &pb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.CustomerService.Login(context.Background(), grpcReq)
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": grpcRes.(*pb.LoginResponse).Token, "message": grpcRes.(*pb.LoginResponse).Message})
}

// @Summary Logout a user
// @Description Logout a user by invalidating their token
// @Tags Customer
// @Accept json
// @Produce json
// @Param Authorization header string true "Token"
// @Success 200
// @Failure 500
// @Router /logout [post]
func Logout(c *gin.Context, grpcClient *grpcclient.GRPCClient, cb *gobreaker.CircuitBreaker) {
	token := c.GetHeader("Authorization")

	grpcReq := &pb.LogoutRequest{
		Token: token,
	}

	grpcRes, err := cb.Execute(func() (interface{}, error) {
		return grpcClient.CustomerService.Logout(context.Background(), grpcReq)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": grpcRes.(*pb.LogoutResponse).Message})
}
