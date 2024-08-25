package main

import (
	"context"
	"testing"

	repository "github.com/m-dehghani/account-service/domain/data"
	"github.com/m-dehghani/account-service/domain/entity"
	"github.com/m-dehghani/account-service/domain/services"

	pb "github.com/m-dehghani/account-service/proto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&entity.Account{}, &entity.Transaction{})
	return db
}

func TestCreateAccount(t *testing.T) {
	db := setupTestDB()
	accountService := services.NewAccountService(repository.NewAccountRepository(db))
	s := &Server{accountService: accountService}

	req := &pb.CreateAccountRequest{Customerid: 1}

	resp, err := s.CreateAccount(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	if resp.Success != true {
		t.Errorf("Error in account creation")
	}

	_, errB := s.BalanceInquiry(context.Background(), &pb.BalanceInquiryRequest{Customerid: 1})
	if errB != nil {
		t.Errorf("Error in account creation")
	}
}

func TestWithdraw(t *testing.T) {
	db := setupTestDB()
	accountService := services.NewAccountService(repository.NewAccountRepository(db))
	s := &Server{accountService: accountService}

	// Create an account first
	s.CreateAccount(context.Background(), &pb.CreateAccountRequest{Customerid: 1})

	_, errD := s.Deposit(context.Background(), &pb.DepositRequest{Customerid: 1, Amount: 1000})

	if errD != nil {
		t.Fatalf("error in deposit")
	}
	req := &pb.WithdrawRequest{Customerid: 1, Amount: 500}
	_, errW := s.Withdraw(context.Background(), req)
	if errW != nil {
		t.Fatalf("Withdraw failed: %v", errW)
	}

	balanceResp, errB := s.BalanceInquiry(context.Background(), &pb.BalanceInquiryRequest{Customerid: 1})

	if errB != nil {
		t.Error(errB.Error())
	}
	if balanceResp.Balance != 500 {
		t.Errorf("Expected new balance to be 500, got %v", balanceResp.Balance)
	}
}

func TestDeposit(t *testing.T) {
	db := setupTestDB()
	accountService := services.NewAccountService(repository.NewAccountRepository(db))
	s := &Server{accountService: accountService}

	// Create an account first
	s.CreateAccount(context.Background(), &pb.CreateAccountRequest{Customerid: 1})

	req := &pb.DepositRequest{Customerid: 1, Amount: 500}
	_, err := s.Deposit(context.Background(), req)
	if err != nil {
		t.Fatalf("Deposit failed: %v", err)
	}
	req2 := &pb.DepositRequest{Customerid: 1, Amount: 1000}
	_, err2 := s.Deposit(context.Background(), req2)
	if err2 != nil {
		t.Fatalf("Deposit failed: %v", err2)
	}

	balanceResp, errB := s.BalanceInquiry(context.Background(), &pb.BalanceInquiryRequest{Customerid: 1})

	if errB != nil {
		t.Error(errB.Error())
	}

	if balanceResp.Balance != 2000 {
		t.Errorf("Expected new balance to be 2000, got %v", balanceResp.Balance)
	}
}

func TestBalanceInquiry(t *testing.T) {
	db := setupTestDB()
	accountService := services.NewAccountService(repository.NewAccountRepository(db))
	s := &Server{accountService: accountService}

	// Create an account first
	s.CreateAccount(context.Background(), &pb.CreateAccountRequest{Customerid: 1})

	req2 := &pb.DepositRequest{Customerid: 1, Amount: 1000}
	_, err2 := s.Deposit(context.Background(), req2)
	if err2 != nil {
		t.Fatalf("Deposit failed: %v", err2)
	}

	req := &pb.BalanceInquiryRequest{Customerid: 1}
	resp, err := s.BalanceInquiry(context.Background(), req)
	if err != nil {
		t.Fatalf("BalanceInquiry failed: %v", err)
	}

	if resp.Balance != 3000 {
		t.Errorf("Expected balance to be 3000, got %v", resp.Balance)
	}
}

func TestTransactionHistory(t *testing.T) {
	db := setupTestDB()
	accountService := services.NewAccountService(repository.NewAccountRepository(db))
	s := &Server{accountService: accountService}

	// Create an account first
	s.CreateAccount(context.Background(), &pb.CreateAccountRequest{Customerid: 1})

	// Perform some transactions
	s.Deposit(context.Background(), &pb.DepositRequest{Customerid: 1, Amount: 500})
	s.Withdraw(context.Background(), &pb.WithdrawRequest{Customerid: 1, Amount: 200})

	req := &pb.TransactionHistoryRequest{Customerid: 1}
	resp, err := s.TransactionHistory(context.Background(), req)
	if err != nil {
		t.Fatalf("TransactionHistory failed: %v", err)
	}

	if len(resp.Transactions) != 7 {
		t.Errorf("Expected 7 transactions, got %v", len(resp.Transactions))
	}
}
