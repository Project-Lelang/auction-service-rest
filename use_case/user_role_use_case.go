package use_case

import (
	"context"

	"auction-service/constant"
	"auction-service/delivery/dto_response"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"
)

type UserRoleUseCase interface {
	AdminCreate(ctx context.Context, userId string, role string)
	AdminDelete(ctx context.Context, userId string, role string)
}

type userRoleUseCase struct {
	repositoryManager repository.RepositoryManager
}

func NewUserRoleUseCase(repositoryManager repository.RepositoryManager) UserRoleUseCase {
	return &userRoleUseCase{repositoryManager: repositoryManager}
}

func (u *userRoleUseCase) AdminCreate(ctx context.Context, userId string, role string) {
	if !isValidRole(role) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageSystemInvalidRequestPayload))
	}

	// Ensure user exists
	mustGetUser(ctx, u.repositoryManager, userId)

	if err := u.repositoryManager.UserRoleRepository().Insert(ctx, &model.UserRole{
		Id:     util.NewUuid(),
		UserId: userId,
		Role:   role,
	}); err != nil {
		if err == constant.ErrDuplicateData {
			panic(dto_response.NewConflictErrorResponse(constant.LanguageUserRoleAlreadyExist))
		}
		panic(err)
	}
}

func (u *userRoleUseCase) AdminDelete(ctx context.Context, userId string, role string) {
	if !isValidRole(role) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageSystemInvalidRequestPayload))
	}

	_, err := u.repositoryManager.UserRoleRepository().GetByUserIdAndRole(ctx, userId, role)
	if err != nil {
		if err == constant.ErrNoData {
			panic(dto_response.NewNotFoundErrorResponse(constant.LanguageUserRoleNotFound))
		}
		panic(err)
	}

	panicIfErr(u.repositoryManager.UserRoleRepository().Delete(ctx, userId, role))
}

func isValidRole(role string) bool {
	switch role {
	case constant.RoleSuperAdmin, constant.RoleAdmin, constant.RoleBidder, constant.RoleSeller:
		return true
	}
	return false
}
