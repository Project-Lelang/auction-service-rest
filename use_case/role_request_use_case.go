package use_case

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

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
	OwnCreate(ctx context.Context, request dto_request.OwnRoleRequestCreateRequest) model.RoleRequest
	AdminFetch(ctx context.Context, req dto_request.RoleRequestFetchRequest) ([]model.RoleRequest, int64)
	AdminFetchByUserId(ctx context.Context, userId int64) []model.RoleRequest
	Approve(ctx context.Context, id int64)
	Reject(ctx context.Context, id int64, req dto_request.RoleRequestRejectRequest)
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

// ------------------------------------------------------------------ helpers

func (u *roleRequestUseCase) mustLoadUserRoles(ctx context.Context, users []*model.User) {
	rolesLoader := loader.NewUserRolesLoader(u.repositoryManager.UserRoleRepository())
	panicIfErr(util.Await(func(group *errgroup.Group) {
		for _, user := range users {
			group.Go(rolesLoader.UserFn(user))
		}
	}))
}

func (u *roleRequestUseCase) populateUserImageLinks(requests ...*model.RoleRequest) {
	const presignedExpiry = 24 * time.Hour
	mainFs := u.filesystemManager.Main()
	for _, request := range requests {
		if request.User == nil {
			continue
		}
		if request.User.IdentityImagePath != nil && *request.User.IdentityImagePath != "" {
			link := mainFs.PresignedUrl(
				util.GetFilenameFromPath(*request.User.IdentityImagePath),
				*request.User.IdentityImagePath,
				presignedExpiry,
			)
			request.User.IdentityImageLink = &link
		}
		if request.User.SelfieIdentityImagePath != nil && *request.User.SelfieIdentityImagePath != "" {
			link := mainFs.PresignedUrl(
				util.GetFilenameFromPath(*request.User.SelfieIdentityImagePath),
				*request.User.SelfieIdentityImagePath,
				presignedExpiry,
			)
			request.User.SelfieIdentityImageLink = &link
		}
	}
}

// ------------------------------------------------------------------ member operations

// OwnCreate dijalankan saat user log-in mengajukan request role baru (dengan validasi berkas & DB)
func (u *roleRequestUseCase) OwnCreate(ctx context.Context, request dto_request.OwnRoleRequestCreateRequest) model.RoleRequest {
	userClaims := model.MustGetUserCtx(ctx)

	// Load user beserta data roles-nya saat ini
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

			identityPath := strings.Trim(strings.TrimPrefix(*request.IdentityImagePath, "/"), " ")
			selfiePath := strings.Trim(strings.TrimPrefix(*request.SelfieIdentityImagePath, "/"), " ")

			u.MustValidateTemporaryFilePaths([]string{identityPath, selfiePath})

			identityMainPath := fmt.Sprintf("user/%d/identity%s", user.Id, path.Ext(identityPath))
			selfieMainPath := fmt.Sprintf("user/%d/selfie_identity%s", user.Id, path.Ext(selfiePath))

			u.MustCopyFromTmpToMain(ctx, identityPath, identityMainPath)
			u.MustCopyFromTmpToMain(ctx, selfiePath, selfieMainPath)

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
			if request.BankAccountNumber == nil || request.BankAccountName == nil || request.BankName == nil {
				panic(dto_response.NewBadRequestErrorResponse(constant.LanguageRoleRequestMissingSellerInfo))
			}
			if err := u.repositoryManager.UserRepository().UpdateBankAccountInfo(ctx, user.Id, *request.BankAccountNumber, *request.BankAccountName, *request.BankName); err != nil {
				return err
			}
		}

		roleRequest = model.RoleRequest{
			UserId: user.Id,
			Status: constant.RoleRequestStatusRequested,
			Role:   request.Role,
		}
		return u.repositoryManager.RoleRequestRepository().Insert(ctx, &roleRequest)
	}))

	return roleRequest
}

// ------------------------------------------------------------------ admin operations

// AdminFetch dipanggil Admin untuk melihat daftar request (Mendukung join user sesuai format data lengkap)
func (u *roleRequestUseCase) AdminFetch(ctx context.Context, req dto_request.RoleRequestFetchRequest) ([]model.RoleRequest, int64) {
	option := model.RoleRequestQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(req.Page, req.Limit, nil),
		Status:      req.Status,
		Role:        req.Role,
	}

	requests, err := u.repositoryManager.RoleRequestRepository().Fetch(ctx, option)
	panicIfErr(err)
	u.populateUserImageLinks(util.SliceValueToSlicePointer(requests)...)

	// Total count dihitung dinamis menggunakan method repository Count demi keandalan pagination
	total, err := u.repositoryManager.RoleRequestRepository().Count(ctx, option)
	panicIfErr(err)

	return requests, total
}

// AdminFetchByUserId dipanggil Admin untuk memeriksa riwayat pengajuan dari 1 user spesifik
func (u *roleRequestUseCase) AdminFetchByUserId(ctx context.Context, userId int64) []model.RoleRequest {
	res, err := u.repositoryManager.RoleRequestRepository().Fetch(ctx, model.RoleRequestQueryOption{
		UserId: &userId,
	})
	panicIfErr(err)
	u.populateUserImageLinks(util.SliceValueToSlicePointer(res)...)
	return res
}

// Approve digunakan oleh Admin untuk menerima permintaan role dan meng-inject role baru ke user terkait
func (u *roleRequestUseCase) Approve(ctx context.Context, id int64) {
	req, err := u.repositoryManager.RoleRequestRepository().GetById(ctx, id)
	panicIfErr(err)

	panicIfErr(u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		// 1. Perbarui status entitas utama menjadi APPROVED
		if err := u.repositoryManager.RoleRequestRepository().UpdateStatus(ctx, id, "APPROVED", nil); err != nil {
			return err
		}

		// 2. Daftarkan berkas role baru ke tabel user_roles
		userRole := &model.UserRole{
			UserId: req.UserId,
			Role:   req.Role,
		}
		return u.repositoryManager.UserRoleRepository().Insert(ctx, userRole)
	}))
}

// Reject digunakan oleh Admin untuk menolak permintaan dengan menyertakan alasan wajib
func (u *roleRequestUseCase) Reject(ctx context.Context, id int64, req dto_request.RoleRequestRejectRequest) {
	roleRequest, err := u.repositoryManager.RoleRequestRepository().GetById(ctx, id)
	panicIfRepositoryError(err, constant.LanguageRoleRequestNotFound)

	panicIfErr(u.repositoryManager.RoleRequestRepository().UpdateStatus(ctx, id, constant.RoleRequestStatusRejected, &req.Message))

	insertUserNotification(
		ctx,
		u.repositoryManager,
		roleRequest.UserId,
		"Role request rejected",
		req.Message,
		"ROLE_REQUEST_REJECTED",
		&id,
	)
}
