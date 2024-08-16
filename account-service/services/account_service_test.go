package services

import (
	"context"
	"testing"
	"time"

	pb "account-service/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open gorm database: %v", err)
	}

	return gormDB, mock
}

func TestWithdraw(t *testing.T) {
	db, mock := setupTestDB(t)
	DB, _ := db.DB()
	defer DB.Close()

	accountService := NewAccountService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE customer_id = \$1 FOR UPDATE`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "balance"}).AddRow(1, 1, 100.0))
	mock.ExpectExec(`UPDATE "accounts" SET "balance"=\$1 WHERE "id" = \$2`).WithArgs(50.0, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "transactions" \("customer_id","type","amount","date"\) VALUES \(\$1,\$2,\$3,\$4\)`).WithArgs(1, "withdraw", 50.0, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := &pb.WithdrawRequest{Customerid: 1, Amount: 50.0}
	res, err := accountService.Withdraw(context.Background(), req)

	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "withdraw successful", res.Message)
}

func TestDeposit(t *testing.T) {
	db, mock := setupTestDB(t)
	DB, _ := db.DB()
	defer DB.Close()

	accountService := NewAccountService(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE customer_id = \$1 FOR UPDATE`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "balance"}).AddRow(1, 1, 100.0))
	mock.ExpectExec(`UPDATE "accounts" SET "balance"=\$1 WHERE "id" = \$2`).WithArgs(150.0, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "transactions" \("customer_id","type","amount","date"\) VALUES \(\$1,\$2,\$3,\$4\)`).WithArgs(1, "deposit", 50.0, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := &pb.DepositRequest{Customerid: 1, Amount: 50.0}
	res, err := accountService.Deposit(context.Background(), req)

	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "deposit successful", res.Message)
}

func TestBalanceInquiry(t *testing.T) {
	db, mock := setupTestDB(t)
	DB, _ := db.DB()
	defer DB.Close()

	accountService := NewAccountService(db)

	mock.ExpectQuery(`SELECT \* FROM "accounts" WHERE customer_id = \$1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "balance"}).AddRow(1, 1, 100.0))

	req := &pb.BalanceInquiryRequest{Customerid: 1}
	res, err := accountService.BalanceInquiry(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, 100.0, res.Balance)
	assert.Equal(t, "balance inquiry successful", res.Message)
}

func TestTransactionHistory(t *testing.T) {
	db, mock := setupTestDB(t)
	DB, _ := db.DB()
	defer DB.Close()

	accountService := NewAccountService(db)

	mock.ExpectQuery(`SELECT \* FROM "transactions" WHERE customer_id = \$1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "type", "amount", "date"}).AddRow(1, 1, "deposit", 50.0, time.Now()).AddRow(2, 1, "withdraw", 30.0, time.Now()))

	req := &pb.TransactionHistoryRequest{Customerid: 1}
	res, err := accountService.TransactionHistory(context.Background(), req)

	assert.NoError(t, err)
	assert.Len(t, res.Transactions, 2)
	assert.Equal(t, "transaction history retrieved", res.Message)
}
