package use_case

import (
	"context"
	"fmt"
	"path"
	"strings"

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
	Fetch(ctx context.Context, req dto_request.RoleRequestFetchRequest) ([]model.RoleRequest, int64)
	FetchByUserId(ctx context.Context, userId string) []model.RoleRequest
	Approve(ctx context.Context, id int64)
	Reject(ctx context.Context, id int64, req dto_request.RoleRequestRejectRequest)
	UserReRequest(ctx context.Context, id int64) model.RoleRequest
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

			identityMainPath := fmt.Sprintf("user/%s/identity%s", user.Id, path.Ext(identityPath))
			selfieMainPath := fmt.Sprintf("user/%s/selfie_identity%s", user.Id, path.Ext(selfiePath))

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
			if request.BankAccountNumber == nil {
				panic(dto_response.NewBadRequestErrorResponse(constant.LanguageRoleRequestMissingSellerInfo))
			}
			if err := u.repositoryManager.UserRepository().UpdateBankAccountNumber(ctx, user.Id, *request.BankAccountNumber); err != nil {
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

// UserReRequest mengajukan kembali request yang sebelumnya ditolak oleh admin
func (u *roleRequestUseCase) UserReRequest(ctx context.Context, id int64) model.RoleRequest {
	userClaims := model.MustGetUserCtx(ctx)

	// 1. Ambil data request lama dari database
	rr, err := u.repositoryManager.RoleRequestRepository().GetById(ctx, id)
	panicIfErr(err)

	// 2. Validasi kepemilikan: Pastikan request ini memang milik user yang sedang login
	if rr.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse("unauthorized: you are not allowed to update this request"))
	}

	// 3. Validasi status: Hanya request bermasalah (REJECTED) yang bisa diajukan kembali
	if rr.Status != "REJECTED" {
		panic(dto_response.NewBadRequestErrorResponse("bad request: only rejected requests can be updated"))
	}

	// 4. Ubah status kembali menjadi REQUESTED dan bersihkan alasan penolakan (message) lama dari admin
	rr.Status = constant.RoleRequestStatusRequested
	rr.Message = nil

	// 5. Simpan perubahan ke DB menggunakan UpdateStatusUser bawaan repository
	panicIfErr(u.repositoryManager.RoleRequestRepository().UpdateStatusUser(ctx, rr))

	return *rr
}

// ------------------------------------------------------------------ admin operations

// Fetch dipanggil Admin untuk melihat daftar request (Mendukung join user sesuai format data lengkap)
func (u *roleRequestUseCase) Fetch(ctx context.Context, req dto_request.RoleRequestFetchRequest) ([]model.RoleRequest, int64) {
	option := model.RoleRequestQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(req.Page, req.Limit, nil),
		Status:      req.Status,
		Role:        req.Role,
	}

	requests, err := u.repositoryManager.RoleRequestRepository().Fetch(ctx, option)
	panicIfErr(err)

	// Total count dihitung dinamis menggunakan method repository Count demi keandalan pagination
	total, err := u.repositoryManager.RoleRequestRepository().Count(ctx, option)
	panicIfErr(err)

	return requests, total
}

// FetchByUserId dipanggil Admin untuk memeriksa riwayat pengajuan dari 1 user spesifik
func (u *roleRequestUseCase) FetchByUserId(ctx context.Context, userId string) []model.RoleRequest {
	res, err := u.repositoryManager.RoleRequestRepository().Fetch(ctx, model.RoleRequestQueryOption{
		UserId: &userId,
	})
	panicIfErr(err)
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
			Id:     util.NewUuid(),
			UserId: req.UserId,
			Role:   req.Role,
		}
		return u.repositoryManager.UserRoleRepository().Insert(ctx, userRole)
	}))
}

// Reject digunakan oleh Admin untuk menolak permintaan dengan menyertakan alasan wajib
func (u *roleRequestUseCase) Reject(ctx context.Context, id int64, req dto_request.RoleRequestRejectRequest) {
	panicIfErr(u.repositoryManager.RoleRequestRepository().UpdateStatus(ctx, id, "REJECTED", &req.Message))
}
