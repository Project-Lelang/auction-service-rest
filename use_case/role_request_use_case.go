package use_case

import (
	"context"
	"fmt"
	"path"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	internalFilesystem "auction-service/internal/filesystem"
	"auction-service/loader"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"

	"golang.org/x/sync/errgroup"
)

type RoleRequestUseCase interface {
	CreateOwn(ctx context.Context, request dto_request.OwnRoleRequestCreateRequest) model.RoleRequest
}

type roleRequestUseCase struct {
	BaseFileUseCase
	repositoryManager repository.RepositoryManager
	filesystemManager internalFilesystem.FilesystemManager
}

func NewRoleRequestUseCase(repositoryManager repository.RepositoryManager, filesystemManager internalFilesystem.FilesystemManager) RoleRequestUseCase {
	return &roleRequestUseCase{
		BaseFileUseCase:   NewBaseFileUseCase(filesystemManager.Main(), filesystemManager.Tmp()),
		repositoryManager: repositoryManager,
		filesystemManager: filesystemManager,
	}
}

func (u *roleRequestUseCase) mustLoadUserRoles(ctx context.Context, users []*model.User) {
	rolesLoader := loader.NewUserRolesLoader(u.repositoryManager.UserRoleRepository())
	panicIfErr(util.Await(func(group *errgroup.Group) {
		for _, user := range users {
			group.Go(rolesLoader.UserFn(user))
		}
	}))
}

func (u *roleRequestUseCase) CreateOwn(ctx context.Context, request dto_request.OwnRoleRequestCreateRequest) model.RoleRequest {
	userClaims := model.MustGetUserCtx(ctx)

	// load user with roles
	user := mustGetUser(ctx, u.repositoryManager, userClaims.UserId)
	u.mustLoadUserRoles(ctx, []*model.User{&user})

	var roleRequest model.RoleRequest

	panicIfErr(u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		switch request.Role {
		case constant.RoleRequestRoleBidder:
			for _, r := range user.Roles {
				if r.Role == constant.RoleBidder {
					panic(dto_response.NewBadRequestErrorResponse(constant.LanguageRoleRequestAlreadyHaveRole))
				}
			}
			existing, err := u.repositoryManager.RoleRequestRepository().GetPendingByUserIdAndRole(ctx, user.Id, constant.RoleRequestRoleBidder)
			if err != nil && err != constant.ErrNoData {
				return err
			}
			if existing != nil {
				panic(dto_response.NewBadRequestErrorResponse(constant.LanguageRoleRequestPendingExists))
			}
			if request.Nik == nil || request.IdentityImagePath == nil || request.SelfieIdentityImagePath == nil {
				panic(dto_response.NewBadRequestErrorResponse(constant.LanguageRoleRequestMissingBidderInfo))
			}
			u.MustValidateTemporaryFilePaths([]string{*request.IdentityImagePath, *request.SelfieIdentityImagePath})
			identityMainPath := fmt.Sprintf("user/%s/identity%s", user.Id, path.Ext(*request.IdentityImagePath))
			selfieMainPath := fmt.Sprintf("user/%s/selfie_identity%s", user.Id, path.Ext(*request.SelfieIdentityImagePath))
			u.MustCopyFromTmpToMain(ctx, *request.IdentityImagePath, identityMainPath)
			u.MustCopyFromTmpToMain(ctx, *request.SelfieIdentityImagePath, selfieMainPath)
			if err := u.repositoryManager.UserRepository().UpdateIdentityInfo(ctx, user.Id, *request.Nik, identityMainPath, selfieMainPath); err != nil {
				return err
			}

		case constant.RoleRequestRoleSeller:
			isBidder := false
			for _, r := range user.Roles {
				if r.Role == constant.RoleBidder {
					isBidder = true
				}
				if r.Role == constant.RoleSeller {
					panic(dto_response.NewBadRequestErrorResponse(constant.LanguageRoleRequestAlreadyHaveRole))
				}
			}
			if !isBidder {
				panic(dto_response.NewBadRequestErrorResponse(constant.LanguageRoleRequestPrerequisiteNotMet))
			}
			existing, err := u.repositoryManager.RoleRequestRepository().GetPendingByUserIdAndRole(ctx, user.Id, constant.RoleRequestRoleSeller)
			if err != nil && err != constant.ErrNoData {
				return err
			}
			if existing != nil {
				panic(dto_response.NewBadRequestErrorResponse(constant.LanguageRoleRequestPendingExists))
			}
			if request.BankAccountNumber == nil {
				panic(dto_response.NewBadRequestErrorResponse(constant.LanguageRoleRequestMissingSellerInfo))
			}
			if err := u.repositoryManager.UserRepository().UpdateBankAccountNumber(ctx, user.Id, *request.BankAccountNumber); err != nil {
				return err
			}
		}

		roleRequest = model.RoleRequest{
			Id:     util.NewUuid(),
			UserId: user.Id,
			Status: constant.RoleRequestStatusRequested,
			Role:   request.Role,
		}
		return u.repositoryManager.RoleRequestRepository().Insert(ctx, &roleRequest)
	}))

	return roleRequest
}
