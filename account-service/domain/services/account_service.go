package services

import (
	"context"
	"fmt"
	"time"

	repository "github.com/m-dehghani/account-service/domain/data"
	"github.com/m-dehghani/account-service/domain/entity"
	pb "github.com/m-dehghani/account-service/proto"
)

type AccountService struct {
	repo repository.AccountRepository
}

func NewAccountService(repo repository.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	account := entity.Account{
		CustomerID: uint(req.Customerid),
		Balance:    0.0,
	}
	if err := s.repo.CreateAccount(ctx, &account); err != nil {
		return &pb.CreateAccountResponse{Success: false, Message: "failed to create account"}, nil
	}
	return &pb.CreateAccountResponse{Success: true, Message: "account created successfully"}, nil
}

func (s *AccountService) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*pb.WithdrawResponse, error) {
	tx := s.repo.Begin()
	if tx.Error != nil {
		return &pb.WithdrawResponse{Success: false, Message: "failed to start transaction"}, nil
	}

	account, err := s.repo.GetAccountByCustomerID(ctx, uint(req.Customerid))
	if err != nil {
		tx.Rollback()
		return &pb.WithdrawResponse{Success: false, Message: "account not found"}, nil
	}

	if account.Balance < req.Amount {
		tx.Rollback()
		return &pb.WithdrawResponse{Success: false, Message: "insufficient funds"}, nil
	}

	account.Balance -= req.Amount
	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		tx.Rollback()
		return &pb.WithdrawResponse{Success: false, Message: "failed to update balance"}, nil
	}

	transaction := entity.Transaction{
		CustomerID: uint(req.Customerid),
		Type:       "withdraw",
		Amount:     req.Amount,
		Date:       time.Now(),
	}
	if err := s.repo.CreateTransaction(ctx, &transaction); err != nil {
		tx.Rollback()
		return &pb.WithdrawResponse{Success: false, Message: "failed to record transaction"}, nil
	}

	if err := tx.Commit().Error; err != nil {
		return &pb.WithdrawResponse{Success: false, Message: "failed to commit transaction"}, nil
	}

	// Send event
	event := "Withdraw successful for customer ID: " + string(rune(req.Customerid))
	fmt.Println(event)

	return &pb.WithdrawResponse{Success: true, Message: "withdraw successful"}, nil
}

func (s *AccountService) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	tx := s.repo.Begin()
	if tx.Error != nil {
		return &pb.DepositResponse{Success: false, Message: "failed to start transaction"}, nil
	}

	account, err := s.repo.GetAccountByCustomerID(ctx, uint(req.Customerid))
	if err != nil {
		tx.Rollback()
		return &pb.DepositResponse{Success: false, Message: "account not found"}, nil
	}

	account.Balance += req.Amount
	if err := s.repo.UpdateAccount(ctx, account); err != nil {
		tx.Rollback()
		return &pb.DepositResponse{Success: false, Message: "failed to update balance"}, nil
	}

	transaction := entity.Transaction{
		CustomerID: uint(req.Customerid),
		Type:       "deposit",
		Amount:     req.Amount,
		Date:       time.Now(),
	}
	if err := s.repo.CreateTransaction(ctx, &transaction); err != nil {
		tx.Rollback()
		return &pb.DepositResponse{Success: false, Message: "failed to record transaction"}, nil
	}

	if err := tx.Commit().Error; err != nil {
		return &pb.DepositResponse{Success: false, Message: "failed to commit transaction"}, nil
	}

	// Send event
	event := "Deposit successful for customer ID: " + fmt.Sprint(req.Customerid)
	fmt.Println(event)

	return &pb.DepositResponse{Success: true, Message: "deposit successful"}, nil
}

func (s *AccountService) BalanceInquiry(ctx context.Context, req *pb.BalanceInquiryRequest) (*pb.BalanceInquiryResponse, error) {
	account, err := s.repo.GetAccountByCustomerID(ctx, uint(req.Customerid))
	if err != nil {
		return &pb.BalanceInquiryResponse{Balance: 0, Message: "account not found"}, nil
	}

	return &pb.BalanceInquiryResponse{Balance: account.Balance, Message: "balance inquiry successful"}, nil
}

func (s *AccountService) TransactionHistory(ctx context.Context, req *pb.TransactionHistoryRequest) (*pb.TransactionHistoryResponse, error) {
	transactions, err := s.repo.GetTransactionsByCustomerID(ctx, uint(req.Customerid))
	if err != nil {
		return &pb.TransactionHistoryResponse{Transactions: nil, Message: "no transactions found"}, nil
	}

	var grpcTransactions []*pb.Transaction
	for _, t := range transactions {
		grpcTransactions = append(grpcTransactions, &pb.Transaction{
			Id:         uint32(t.ID),
			Customerid: uint32(t.CustomerID),
			Type:       t.Type,
			Amount:     t.Amount,
			Date:       t.Date.Format(time.RFC3339),
		})
	}

	return &pb.TransactionHistoryResponse{Transactions: grpcTransactions, Message: "transaction history retrieved"}, nil
}
