package repository

import (
	"context"
	"database/sql"
	"fmt"

	"auction-service/model"

	"github.com/jmoiron/sqlx"
)

type RepositoryManager interface {
	UserRepository() UserRepository
	UserRoleRepository() UserRoleRepository
	OtpRepository() OtpRepository
	ProductRepository() ProductRepository
	ProductStatusHistoryRepository() ProductStatusHistoryRepository
	RoleRequestRepository() RoleRequestRepository
	WithdrawalRequestRepository() WithdrawalRequestRepository
	AuctionRepository() AuctionRepository
	AuctionBidRepository() AuctionBidRepository
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	PaymentMethodRepository() PaymentMethodRepository
}

type repositoryManager struct {
	db                             *sqlx.DB
	userRepository                 UserRepository
	userRoleRepository             UserRoleRepository
	otpRepository                  OtpRepository
	productRepository              ProductRepository
	productStatusHistoryRepository ProductStatusHistoryRepository
	roleRequestRepository          RoleRequestRepository
	withdrawalRequestRepository    WithdrawalRequestRepository
	auctionRepository              AuctionRepository
	auctionBidRepository           AuctionBidRepository
	paymentMethodRepository        PaymentMethodRepository
}

func NewRepositoryManager(
	db *sqlx.DB,
	userRepository UserRepository,
	userRoleRepository UserRoleRepository,
	otpRepository OtpRepository,
	productRepository ProductRepository,
	productStatusHistoryRepository ProductStatusHistoryRepository,
	roleRequestRepository RoleRequestRepository,
	withdrawalRequestRepository WithdrawalRequestRepository,
	auctionRepository AuctionRepository,
	auctionBidRepository AuctionBidRepository,
	paymentMethodRepository PaymentMethodRepository,
) RepositoryManager {
	return &repositoryManager{
		db:                             db,
		userRepository:                 userRepository,
		userRoleRepository:             userRoleRepository,
		otpRepository:                  otpRepository,
		productRepository:              productRepository,
		productStatusHistoryRepository: productStatusHistoryRepository,
		roleRequestRepository:          roleRequestRepository,
		withdrawalRequestRepository:    withdrawalRequestRepository,
		auctionRepository:              auctionRepository,
		auctionBidRepository:           auctionBidRepository,
		paymentMethodRepository:        paymentMethodRepository,
	}
}

func (r *repositoryManager) Transaction(
	ctx context.Context,
	fn func(ctx context.Context) error,
) (err error) {
	var tx *sqlx.Tx

	defer func() {
		if err != nil && tx != nil {
			if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
				err = fmt.Errorf("%v\nrollback err: %v", err, rbErr)
			}
		}
	}()

	tx, err = r.db.BeginTxx(ctx, nil)
	if err != nil {
		return
	}

	ctx, err = model.SetDbtxCtx(ctx, tx)
	if err != nil {
		return
	}

	if err = fn(ctx); err != nil {
		return
	}

	return tx.Commit()
}

func (r *repositoryManager) UserRepository() UserRepository {
	return r.userRepository
}

func (r *repositoryManager) UserRoleRepository() UserRoleRepository {
	return r.userRoleRepository
}

func (r *repositoryManager) OtpRepository() OtpRepository {
	return r.otpRepository
}

func (r *repositoryManager) ProductRepository() ProductRepository {
	return r.productRepository
}

func (r *repositoryManager) ProductStatusHistoryRepository() ProductStatusHistoryRepository {
	return r.productStatusHistoryRepository
}

func (r *repositoryManager) RoleRequestRepository() RoleRequestRepository {
	return r.roleRequestRepository
}

func (r *repositoryManager) WithdrawalRequestRepository() WithdrawalRequestRepository {
	return r.withdrawalRequestRepository
}

func (r *repositoryManager) AuctionRepository() AuctionRepository {
	return r.auctionRepository
}

func (r *repositoryManager) AuctionBidRepository() AuctionBidRepository {
	return r.auctionBidRepository
}

func (r *repositoryManager) PaymentMethodRepository() PaymentMethodRepository {
	return r.paymentMethodRepository
}
