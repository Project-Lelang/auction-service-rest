package use_case

import (
	"context"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/loader"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"

	"golang.org/x/sync/errgroup"
)

type WithdrawalRequestUseCase interface {
	CreateOwn(ctx context.Context, request dto_request.OwnWithdrawalRequestCreateRequest) model.WithdrawalRequest
}

type withdrawalRequestUseCase struct {
	repositoryManager repository.RepositoryManager
}

func NewWithdrawalRequestUseCase(repositoryManager repository.RepositoryManager) WithdrawalRequestUseCase {
	return &withdrawalRequestUseCase{repositoryManager: repositoryManager}
}

func (u *withdrawalRequestUseCase) mustLoadUserRoles(ctx context.Context, users []*model.User) {
	rolesLoader := loader.NewUserRolesLoader(u.repositoryManager.UserRoleRepository())
	panicIfErr(util.Await(func(group *errgroup.Group) {
		for _, user := range users {
			group.Go(rolesLoader.UserFn(user))
		}
	}))
}

func (u *withdrawalRequestUseCase) CreateOwn(ctx context.Context, request dto_request.OwnWithdrawalRequestCreateRequest) model.WithdrawalRequest {
	userClaims := model.MustGetUserCtx(ctx)

	// load user with roles
	user := mustGetUser(ctx, u.repositoryManager, userClaims.UserId)
	u.mustLoadUserRoles(ctx, []*model.User{&user})

	// must be BIDDER or SELLER
	hasBidderOrSeller := false
	for _, r := range user.Roles {
		if r.Role == constant.RoleBidder || r.Role == constant.RoleSeller {
			hasBidderOrSeller = true
			break
		}
	}
	if !hasBidderOrSeller {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	// must have bank account
	if user.BankAccountNumber == nil || *user.BankAccountNumber == "" {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageWithdrawalRequestNoBankAccount))
	}

	// must have sufficient balance
	if user.Balance < request.Amount {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageWithdrawalRequestInsufficientBalance))
	}

	withdrawalRequest := model.WithdrawalRequest{
		Id:     util.NewUuid(),
		UserId: user.Id,
		Amount: request.Amount,
		Status: constant.WithdrawalRequestStatusRequested,
	}
	panicIfErr(u.repositoryManager.WithdrawalRequestRepository().Insert(ctx, &withdrawalRequest))

	return withdrawalRequest
}
