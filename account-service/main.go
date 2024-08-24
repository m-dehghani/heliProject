package main

import (
	"context"
	"log"
	"net"
	"os"

	repository "github.com/m-dehghani/account-service/domain/data"
	"github.com/m-dehghani/account-service/domain/entity"
	"github.com/m-dehghani/account-service/domain/services"
	pb "github.com/m-dehghani/account-service/proto"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type server struct {
	accountService *services.AccountService
}

// CreateAccount implements proto.AccountServiceServer.
func (s *server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	return s.accountService.CreateAccount(ctx, req)
}

func (s *server) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*pb.WithdrawResponse, error) {
	return s.accountService.Withdraw(ctx, req)
}

func (s *server) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	return s.accountService.Deposit(ctx, req)
}

func (s *server) BalanceInquiry(ctx context.Context, req *pb.BalanceInquiryRequest) (*pb.BalanceInquiryResponse, error) {
	return s.accountService.BalanceInquiry(ctx, req)
}

func (s *server) TransactionHistory(ctx context.Context, req *pb.TransactionHistoryRequest) (*pb.TransactionHistoryResponse, error) {
	return s.accountService.TransactionHistory(ctx, req)
}

func main() {
	dsn := "host=" + os.Getenv("POSTGRES_HOST") + " user=" + os.Getenv("POSTGRES_USER") + " password=" + os.Getenv("POSTGRES_PASSWORD") + " dbname=" + os.Getenv("POSTGRES_DB") + " port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&entity.Account{}, &entity.Transaction{})

	accountService := services.NewAccountService(repository.NewAccountRepository(db))

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterAccountServiceServer(s, &server{accountService: accountService})

	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
