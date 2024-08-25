package account

import (
	"context"

	pb "github.com/m-dehghani/gateway-service/proto"
)

type AccountService struct {
	client pb.AccountServiceClient
}

func NewAccountService(client pb.AccountServiceClient) *AccountService {
	return &AccountService{client: client}
}

func (s *AccountService) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	return s.client.CreateAccount(ctx, req)
}

func (s *AccountService) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	return s.client.Deposit(ctx, req)
}

func (s *AccountService) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*pb.WithdrawResponse, error) {
	return s.client.Withdraw(ctx, req)
}

func (s *AccountService) BalanceInquiry(ctx context.Context, req *pb.BalanceInquiryRequest) (*pb.BalanceInquiryResponse, error) {
	return s.client.BalanceInquiry(ctx, req)
}

func (s *AccountService) TransactionHistory(ctx context.Context, req *pb.TransactionHistoryRequest) (*pb.TransactionHistoryResponse, error) {
	return s.client.TransactionHistory(ctx, req)
}
