package services

import (
	"context"
	"fmt"
	"time"

	pb "account-service/proto"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountService struct {
	db *gorm.DB
	// mq *amqp.Channel
}

type Account struct {
	ID         uint `gorm:"primaryKey"`
	CustomerID uint
	Balance    float64
}

type Transaction struct {
	ID         uint `gorm:"primaryKey"`
	CustomerID uint
	Type       string
	Amount     float64
	Date       time.Time
}

func NewAccountService(db *gorm.DB) *AccountService {
	return &AccountService{db: db}
}

func (s *AccountService) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	account := Account{
		CustomerID: uint(req.Customerid),
		Balance:    0.0,
	}
	if err := s.db.Create(&account).Error; err != nil {
		return &pb.CreateAccountResponse{Success: false, Message: "failed to create account"}, nil
	}
	return &pb.CreateAccountResponse{Success: true, Message: "account created successfully"}, nil
}

func (s *AccountService) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*pb.WithdrawResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return &pb.WithdrawResponse{Success: false, Message: "failed to start transaction"}, nil
	}

	var account Account
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("customer_id = ?", req.Customerid).First(&account).Error; err != nil {
		tx.Rollback()
		return &pb.WithdrawResponse{Success: false, Message: "account not found"}, nil
	}

	if account.Balance < req.Amount {
		tx.Rollback()
		return &pb.WithdrawResponse{Success: false, Message: "insufficient funds"}, nil
	}

	account.Balance -= req.Amount
	if err := tx.Save(&account).Error; err != nil {
		tx.Rollback()
		return &pb.WithdrawResponse{Success: false, Message: "failed to update balance"}, nil
	}

	transaction := Transaction{
		CustomerID: uint(req.Customerid),
		Type:       "withdraw",
		Amount:     req.Amount,
		Date:       time.Now(),
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return &pb.WithdrawResponse{Success: false, Message: "failed to record transaction"}, nil
	}

	if err := tx.Commit().Error; err != nil {
		return &pb.WithdrawResponse{Success: false, Message: "failed to commit transaction"}, nil
	}

	// Send event
	event := "Withdraw successful for customer ID: " + string(rune(req.Customerid))
	fmt.Println(event)
	// err := s.mq.Publish(
	// 	"",               // exchange
	// 	"withdraw_queue", // routing key
	// 	false,            // mandatory
	// 	false,            // immediate
	// 	amqp.Publishing{
	// 		ContentType: "text/plain",
	// 		Body:        []byte(event),
	// 	})
	// if err != nil {
	// 	return &pb.WithdrawResponse{Success: false, Message: "failed to publish event"}, nil
	// }

	return &pb.WithdrawResponse{Success: true, Message: "withdraw successful"}, nil
}

func (s *AccountService) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return &pb.DepositResponse{Success: false, Message: "failed to start transaction"}, nil
	}

	var account Account
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("customer_id = ?", req.Customerid).First(&account).Error; err != nil {
		tx.Rollback()
		return &pb.DepositResponse{Success: false, Message: "account not found"}, nil
	}

	account.Balance += req.Amount
	if err := tx.Save(&account).Error; err != nil {
		tx.Rollback()
		return &pb.DepositResponse{Success: false, Message: "failed to update balance"}, nil
	}

	transaction := Transaction{
		CustomerID: uint(req.Customerid),
		Type:       "deposit",
		Amount:     req.Amount,
		Date:       time.Now(),
	}
	if err := tx.Create(&transaction).Error; err != nil {
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
	var account Account
	if err := s.db.Where("customer_id = ?", req.Customerid).First(&account).Error; err != nil {
		return &pb.BalanceInquiryResponse{Balance: 0, Message: "account not found"}, nil
	}

	return &pb.BalanceInquiryResponse{Balance: account.Balance, Message: "balance inquiry successful"}, nil
}

func (s *AccountService) TransactionHistory(ctx context.Context, req *pb.TransactionHistoryRequest) (*pb.TransactionHistoryResponse, error) {
	var transactions []Transaction
	if err := s.db.Where("customer_id = ?", req.Customerid).Find(&transactions).Error; err != nil {
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
