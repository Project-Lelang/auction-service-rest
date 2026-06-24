package use_case

import (
	"context"
	"time"

	"auction-service/constant"
	"auction-service/data_type"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	internalFilesystem "auction-service/internal/filesystem"
	"auction-service/loader"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"

	"golang.org/x/sync/errgroup"
)

type UserUseCase interface {
	// create
	AdminCreate(ctx context.Context, request dto_request.AdminUserCreateRequest)

	// read
	AdminFetch(ctx context.Context, request dto_request.AdminUserFetchRequest) ([]model.User, int64)
	AdminGet(ctx context.Context, request dto_request.AdminUserGetRequest) model.User
	OwnGet(ctx context.Context) model.User

	// update
	OwnUpdate(ctx context.Context, request dto_request.OwnProfileUpdateRequest) model.User

	// delete
	AdminDelete(ctx context.Context, request dto_request.AdminUserDeleteRequest)
}

type userUseCase struct {
	repositoryManager repository.RepositoryManager
	filesystemManager internalFilesystem.FilesystemManager
}

func NewUserUseCase(repositoryManager repository.RepositoryManager, filesystemManager internalFilesystem.FilesystemManager) UserUseCase {
	return &userUseCase{repositoryManager: repositoryManager, filesystemManager: filesystemManager}
}

type userLoaderParams struct {
	roles bool
}

func (u *userUseCase) mustLoadUserData(_ context.Context, users []*model.User, option userLoaderParams) {
	userRolesLoader := loader.NewUserRolesLoader(u.repositoryManager.UserRoleRepository())

	panicIfErr(util.Await(func(group *errgroup.Group) {
		for _, user := range users {
			if option.roles {
				group.Go(userRolesLoader.UserFn(user))
			}
		}
	}))
}

func (u *userUseCase) populateImageLinks(users ...*model.User) {
	const presignedExpiry = 24 * time.Hour
	mainFs := u.filesystemManager.Main()
	for _, user := range users {
		if user.IdentityImagePath != nil && *user.IdentityImagePath != "" {
			link := mainFs.PresignedUrl(util.GetFilenameFromPath(*user.IdentityImagePath), *user.IdentityImagePath, presignedExpiry)
			user.IdentityImageLink = &link
		}
		if user.SelfieIdentityImagePath != nil && *user.SelfieIdentityImagePath != "" {
			link := mainFs.PresignedUrl(util.GetFilenameFromPath(*user.SelfieIdentityImagePath), *user.SelfieIdentityImagePath, presignedExpiry)
			user.SelfieIdentityImageLink = &link
		}
	}
}

func (u *userUseCase) AdminCreate(ctx context.Context, request dto_request.AdminUserCreateRequest) {
	hashedPassword, hashErr := util.HashPassword(request.Password)
	panicIfErr(hashErr)

	birth, parseErr := time.Parse("2006-01-02", request.Birth)
	if parseErr != nil {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageSystemInvalidRequestPayload))
	}

	user := &model.User{
		Fullname: request.Fullname,
		Email:    normalizeEmail(request.Email),
		Birth:    data_type.NewDateTime(birth),
		Gender:   request.Gender,
		Password: hashedPassword,
	}

	if err := u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		if err := u.repositoryManager.UserRepository().Insert(ctx, user); err != nil {
			return err
		}
		return u.repositoryManager.UserRoleRepository().Insert(ctx, &model.UserRole{
			UserId: user.Id,
			Role:   constant.RoleAdmin,
		})
	}); err != nil {
		if err == constant.ErrDuplicateData {
			panic(dto_response.NewConflictErrorResponse(constant.LanguageUserEmailAlreadyExist))
		}
		panic(err)
	}
}

func (u *userUseCase) AdminFetch(ctx context.Context, request dto_request.AdminUserFetchRequest) ([]model.User, int64) {
	option := model.UserQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(request.Page, request.Limit, model.Sorts(request.Sorts)),
		Role:        request.Role,
	}

	total, err := u.repositoryManager.UserRepository().Count(ctx, option)
	panicIfErr(err)

	users, err := u.repositoryManager.UserRepository().Fetch(ctx, option)
	panicIfErr(err)

	u.mustLoadUserData(ctx, util.SliceValueToSlicePointer(users), userLoaderParams{roles: true})

	return users, total
}

func (u *userUseCase) AdminGet(ctx context.Context, request dto_request.AdminUserGetRequest) model.User {
	user := mustGetUser(ctx, u.repositoryManager, request.UserId)
	u.mustLoadUserData(ctx, []*model.User{&user}, userLoaderParams{roles: true})
	u.populateImageLinks(&user)
	return user
}

func (u *userUseCase) OwnGet(ctx context.Context) model.User {
	userClaims := model.MustGetUserCtx(ctx)
	user := mustGetUser(ctx, u.repositoryManager, userClaims.UserId)
	u.mustLoadUserData(ctx, []*model.User{&user}, userLoaderParams{roles: true})
	u.populateImageLinks(&user)
	return user
}

func (u *userUseCase) OwnUpdate(ctx context.Context, request dto_request.OwnProfileUpdateRequest) model.User {
	userClaims := model.MustGetUserCtx(ctx)

	birth, parseErr := time.Parse("2006-01-02", request.Birth)
	if parseErr != nil {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageSystemInvalidRequestPayload))
	}

	updated, err := u.repositoryManager.UserRepository().Update(
		ctx,
		userClaims.UserId,
		request.Fullname,
		data_type.NewDateTime(birth),
		request.Gender,
	)
	panicIfErr(err)

	return *updated
}

func (u *userUseCase) AdminDelete(ctx context.Context, request dto_request.AdminUserDeleteRequest) {
	mustGetUser(ctx, u.repositoryManager, request.UserId)
	panicIfErr(u.repositoryManager.UserRepository().SoftDelete(ctx, request.UserId))
}
