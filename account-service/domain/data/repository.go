package repository

import (
	"context"

	"github.com/m-dehghani/account-service/domain/entity"

	"gorm.io/gorm"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, account *entity.Account) error
	GetAccountByCustomerID(ctx context.Context, customerID uint) (*entity.Account, error)
	UpdateAccount(ctx context.Context, account *entity.Account) error
	CreateTransaction(ctx context.Context, transaction *entity.Transaction) error
	GetTransactionsByCustomerID(ctx context.Context, customerID uint) ([]entity.Transaction, error)
	Begin() *gorm.DB
}

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) CreateAccount(ctx context.Context, account *entity.Account) error {
	return r.db.Create(account).Error
}

func (r *accountRepository) GetAccountByCustomerID(ctx context.Context, customerID uint) (*entity.Account, error) {
	var account entity.Account
	err := r.db.Where("customer_id = ?", customerID).First(&account).Error
	return &account, err
}

func (r *accountRepository) UpdateAccount(ctx context.Context, account *entity.Account) error {
	return r.db.Save(account).Error
}

func (r *accountRepository) CreateTransaction(ctx context.Context, transaction *entity.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *accountRepository) GetTransactionsByCustomerID(ctx context.Context, customerID uint) ([]entity.Transaction, error) {
	var transactions []entity.Transaction
	err := r.db.Where("customer_id = ?", customerID).Find(&transactions).Error
	return transactions, err
}

func (r *accountRepository) Begin() *gorm.DB {
	return r.db.Begin()
}
