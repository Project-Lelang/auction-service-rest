package use_case

import (
	"context"
	"time"

	"auction-service/constant"
	"auction-service/data_type"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
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
}

func NewUserUseCase(repositoryManager repository.RepositoryManager) UserUseCase {
	return &userUseCase{repositoryManager: repositoryManager}
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

func (u *userUseCase) AdminCreate(ctx context.Context, request dto_request.AdminUserCreateRequest) {
	hashedPassword, hashErr := util.HashPassword(request.Password)
	panicIfErr(hashErr)

	birth, parseErr := time.Parse("2006-01-02", request.Birth)
	if parseErr != nil {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageSystemInvalidRequestPayload))
	}

	user := &model.User{
		Id:       util.NewUuid(),
		Fullname: request.Fullname,
		Phone:    request.Phone,
		Birth:    data_type.NewDateTime(birth),
		Gender:   request.Gender,
		Password: hashedPassword,
	}

	insertErr := u.repositoryManager.UserRepository().Insert(ctx, user)
	if insertErr != nil {
		if insertErr == constant.ErrDuplicateData {
			panic(dto_response.NewConflictErrorResponse(constant.LanguageAuthPhoneAlreadyRegistered))
		}
		panic(insertErr)
	}

	if err := u.repositoryManager.UserRoleRepository().Insert(ctx, &model.UserRole{
		Id:     util.NewUuid(),
		UserId: user.Id,
		Role:   constant.RoleAdmin,
	}); err != nil {
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
	return user
}

func (u *userUseCase) OwnGet(ctx context.Context) model.User {
	userClaims := model.MustGetUserCtx(ctx)
	user := mustGetUser(ctx, u.repositoryManager, userClaims.UserId)
	u.mustLoadUserData(ctx, []*model.User{&user}, userLoaderParams{roles: true})
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
		request.Phone,
		request.Nik,
		data_type.NewDateTime(birth),
		request.Gender,
		request.BankAccountNumber,
	)
	panicIfErr(err)

	return *updated
}

func (u *userUseCase) AdminDelete(ctx context.Context, request dto_request.AdminUserDeleteRequest) {
	mustGetUser(ctx, u.repositoryManager, request.UserId)
	panicIfErr(u.repositoryManager.UserRepository().SoftDelete(ctx, request.UserId))
}
