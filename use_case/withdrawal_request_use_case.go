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
	// USER
	OwnCreate(ctx context.Context, request dto_request.OwnWithdrawalRequestCreateRequest) model.WithdrawalRequest

	// ADMIN
	Fetch(ctx context.Context, req dto_request.WithdrawalRequestFetchRequest) ([]model.WithdrawalRequest, int64)
	FetchByUserId(ctx context.Context, userId string, req dto_request.WithdrawalRequestFetchRequest) ([]model.WithdrawalRequest, int64)
	Complete(ctx context.Context, adminId string, userId string, withdrawalRequestId int64) model.WithdrawalRequest
}

type withdrawalRequestUseCase struct {
	repositoryManager repository.RepositoryManager
}

func NewWithdrawalRequestUseCase(repositoryManager repository.RepositoryManager) WithdrawalRequestUseCase {
	return &withdrawalRequestUseCase{
		repositoryManager: repositoryManager,
	}
}

// ------------------------------------------------------------------ helpers

func (u *withdrawalRequestUseCase) mustLoadUserRoles(ctx context.Context, users []*model.User) {
	rolesLoader := loader.NewUserRolesLoader(u.repositoryManager.UserRoleRepository())
	panicIfErr(util.Await(func(group *errgroup.Group) {
		for _, user := range users {
			group.Go(rolesLoader.UserFn(user))
		}
	}))
}

func (u *withdrawalRequestUseCase) mustLoadUsers(ctx context.Context, requests []model.WithdrawalRequest) {
	if len(requests) == 0 {
		return
	}

	// 1. Kumpulkan semua userId unik dari list requests
	userIdsMap := make(map[string]bool)
	for _, req := range requests {
		if req.UserId != "" {
			userIdsMap[req.UserId] = true
		}
	}

	userIds := make([]string, 0, len(userIdsMap))
	for id := range userIdsMap {
		userIds = append(userIds, id)
	}

	// 2. Tarik data user secara massal dari database (Mengembalikan []model.User)
	users, err := u.repositoryManager.UserRepository().FetchByIds(ctx, userIds)
	panicIfErr(err)

	// 3. Pindahkan ke map untuk mempermudah mapping data kembali
	userMap := make(map[string]model.User)
	for _, user := range users {
		userMap[user.Id] = user
	}

	// 4. Petakan model.User ke model.WithdrawalRequestUser yang diinginkan oleh struct
	for i := range requests {
		if dbUser, exists := userMap[requests[i].UserId]; exists {
			// Lakukan mapping manual di sini agar tipe data COCOK
			requests[i].User = &model.WithdrawalRequestUser{
				Id:                dbUser.Id,
				Fullname:          dbUser.Fullname,
				Phone:             dbUser.Phone,
				BankAccountNumber: dbUser.BankAccountNumber,
				BankName:          dbUser.BankName,
				BankAccountName:   dbUser.BankAccountName,
			}
		}
	}
}

// ------------------------------------------------------------------ member operations

// OwnCreate dijalankan oleh user log-in untuk mengajukan penarikan saldo
func (u *withdrawalRequestUseCase) OwnCreate(ctx context.Context, request dto_request.OwnWithdrawalRequestCreateRequest) model.WithdrawalRequest {
	userClaims := model.MustGetUserCtx(ctx)

	user := mustGetUser(ctx, u.repositoryManager, userClaims.UserId)
	u.mustLoadUserRoles(ctx, []*model.User{&user})

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

	if user.BankAccountNumber == nil || *user.BankAccountNumber == "" {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageWithdrawalRequestNoBankAccount))
	}

	if user.Balance < request.Amount {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageWithdrawalRequestInsufficientBalance))
	}

	withdrawalRequest := model.WithdrawalRequest{
		UserId: user.Id,
		Amount: request.Amount,
		Status: constant.WithdrawalRequestStatusRequested,
	}
	panicIfErr(u.repositoryManager.WithdrawalRequestRepository().Insert(ctx, &withdrawalRequest))

	return withdrawalRequest
}

// ------------------------------------------------------------------ admin operations

// Fetch dipanggil Admin untuk melihat seluruh daftar pengajuan dana (Sesuai kebutuhan endpoint UI Frontend Anda)
func (u *withdrawalRequestUseCase) Fetch(ctx context.Context, req dto_request.WithdrawalRequestFetchRequest) ([]model.WithdrawalRequest, int64) {
	option := model.WithdrawalRequestQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(req.Page, req.Limit, nil),
		Status:      req.Status,
	}

	requests, err := u.repositoryManager.WithdrawalRequestRepository().Fetch(ctx, option)
	panicIfErr(err)

	total, err := u.repositoryManager.WithdrawalRequestRepository().Count(ctx, option)
	panicIfErr(err)

	u.mustLoadUsers(ctx, requests)

	return requests, total
}

// FetchByUserId digunakan Admin jika ingin memfilter history pengajuan berdasarkan 1 user spesifik
func (u *withdrawalRequestUseCase) FetchByUserId(
	ctx context.Context,
	userId string,
	req dto_request.WithdrawalRequestFetchRequest,
) ([]model.WithdrawalRequest, int64) {

	user, err := u.repositoryManager.UserRepository().GetById(ctx, userId)
	panicIfErr(err)
	if user == nil {
		panic(dto_response.NewNotFoundErrorResponse("user not found"))
	}

	option := model.WithdrawalRequestQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(req.Page, req.Limit, nil),
		UserId:      &userId,
		Status:      req.Status,
	}

	res, err := u.repositoryManager.WithdrawalRequestRepository().Fetch(ctx, option)
	panicIfErr(err)

	total, err := u.repositoryManager.WithdrawalRequestRepository().Count(ctx, option)
	panicIfErr(err)

	u.mustLoadUsers(ctx, res)

	return res, total
}

// Complete dijalankan oleh Admin untuk memproses persetujuan pencairan dana
func (u *withdrawalRequestUseCase) Complete(
	ctx context.Context,
	adminId string,
	userId string,
	withdrawalRequestId int64,
) model.WithdrawalRequest {

	var updatedWr *model.WithdrawalRequest

	panicIfErr(u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		user, err := u.repositoryManager.UserRepository().GetById(ctx, userId)
		if err != nil {
			return err
		}
		if user == nil {
			panic(dto_response.NewNotFoundErrorResponse("user not found"))
		}

		wr, err := u.repositoryManager.WithdrawalRequestRepository().GetById(ctx, withdrawalRequestId)
		if err != nil {
			return err
		}
		if wr == nil {
			panic(dto_response.NewNotFoundErrorResponse("withdrawal request not found"))
		}

		if wr.UserId != userId {
			panic(dto_response.NewBadRequestErrorResponse("withdrawal request does not belong to user"))
		}
		if wr.Status == model.WithdrawalRequestStatusCompleted {
			panic(dto_response.NewBadRequestErrorResponse("withdrawal request already completed"))
		}

		if user.Balance < wr.Amount {
			panic(dto_response.NewBadRequestErrorResponse(constant.LanguageWithdrawalRequestInsufficientBalance))
		}

		newBalance := user.Balance - wr.Amount
		if err := u.repositoryManager.UserRepository().UpdateBalance(ctx, user.Id, newBalance); err != nil {
			return err
		}

		wrPayload := map[string]interface{}{
			"status":            model.WithdrawalRequestStatusCompleted,
			"validator_user_id": adminId,
		}
		if err := u.repositoryManager.WithdrawalRequestRepository().Update(ctx, withdrawalRequestId, wrPayload); err != nil {
			return err
		}

		updatedWr, err = u.repositoryManager.WithdrawalRequestRepository().GetById(ctx, withdrawalRequestId)
		return err
	}))

	// Ambil data user untuk dikembalikan lengkap ke response action
	resSlice := []model.WithdrawalRequest{*updatedWr}
	u.mustLoadUsers(ctx, resSlice)

	return resSlice[0]
}
