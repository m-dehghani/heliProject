package customer

import (
	"context"

	pb "github.com/m-dehghani/gateway-service/proto"
)

type CustomerService struct {
	client pb.CustomerServiceClient
}

func NewCustomerService(client pb.CustomerServiceClient) *CustomerService {
	return &CustomerService{client: client}
}

func (s *CustomerService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return s.client.Register(ctx, req)
}

func (s *CustomerService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return s.client.Login(ctx, req)
}

func (s *CustomerService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return s.client.Logout(ctx, req)
}

func (s *CustomerService) VerifyCustomerID(ctx context.Context, req *pb.VerifyCustomerIDRequest) (*pb.VerifyCustomerIDResponse, error) {
	return s.client.VerifyCustomerID(ctx, req)
}
