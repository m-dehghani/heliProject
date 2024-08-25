package grpcclient

import (
	"log"

	"github.com/m-dehghani/gateway-service/models/account"
	"github.com/m-dehghani/gateway-service/models/customer"
	pb "github.com/m-dehghani/gateway-service/proto"
	"google.golang.org/grpc"
)

type GRPCClient struct {
	AccountService  *account.AccountService
	CustomerService *customer.CustomerService
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
		AccountService:  account.NewAccountService(pb.NewAccountServiceClient(accountConn)),
		CustomerService: customer.NewCustomerService(pb.NewCustomerServiceClient(customerConn)),
	}
}
