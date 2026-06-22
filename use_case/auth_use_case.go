package use_case

import (
	"context"
	"time"

	"auction-service/constant"
	"auction-service/data_type"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/global"
	internalJwt "auction-service/internal/jwt"
	"auction-service/loader"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"

	"golang.org/x/sync/errgroup"
)

type AuthUseCase interface {
	// Public
	Login(ctx context.Context, request dto_request.AuthLoginRequest) string
	AdminLogin(ctx context.Context, request dto_request.AdminAuthLoginRequest) string
	Register(ctx context.Context, request dto_request.AuthRegisterRequest)
	SaveFcmToken(ctx context.Context, request dto_request.AuthFcmTokenRequest)
	CreateOtp(ctx context.Context, phone string)
	// Middleware helper
	Parse(ctx context.Context, token string) (*model.UserClaims, error)
}

type authUseCase struct {
	repositoryManager repository.RepositoryManager
	jwt               internalJwt.Jwt
}

func NewAuthUseCase(
	repositoryManager repository.RepositoryManager,
	jwt internalJwt.Jwt,
) AuthUseCase {
	return &authUseCase{
		repositoryManager: repositoryManager,
		jwt:               jwt,
	}
}

func (u *authUseCase) generateToken(user *model.User) string {
	roleStrings := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roleStrings[i] = r.Role
	}

	now := util.CurrentDateTime()
	expiry := time.Duration(global.GetConfig().JwtConfig.AccessTokenExpiryHours) * time.Hour

	token, err := u.jwt.Generate(internalJwt.Payload{
		Id:        user.Id,
		Phone:     user.Phone,
		Roles:     roleStrings,
		CreatedAt: now,
		ExpiredAt: now.Add(expiry),
	})
	panicIfErr(err)
	return token.AccessToken
}

func (u *authUseCase) Login(ctx context.Context, request dto_request.AuthLoginRequest) string {
	user, err := u.repositoryManager.UserRepository().GetByPhone(ctx, request.Phone)
	if err != nil {
		if err == constant.ErrNoData {
			panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuthInvalidCredentials))
		}
		panic(err)
	}

	if !util.CheckPasswordHash(request.Password, user.Password) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuthInvalidCredentials))
	}

	userRolesLoader := loader.NewUserRolesLoader(u.repositoryManager.UserRoleRepository())
	panicIfErr(util.Await(func(group *errgroup.Group) {
		group.Go(userRolesLoader.UserFn(user))
	}))
	return u.generateToken(user)
}

func (u *authUseCase) AdminLogin(ctx context.Context, request dto_request.AdminAuthLoginRequest) string {
	user, err := u.repositoryManager.UserRepository().FindAdminByPhone(ctx, request.Phone)
	if err != nil {
		if err == constant.ErrNoData {
			panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuthInvalidCredentials))
		}
		panic(err)
	}

	if !util.CheckPasswordHash(request.Password, user.Password) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuthInvalidCredentials))
	}

	userRolesLoader := loader.NewUserRolesLoader(u.repositoryManager.UserRoleRepository())
	panicIfErr(util.Await(func(group *errgroup.Group) {
		group.Go(userRolesLoader.UserFn(user))
	}))
	return u.generateToken(user)
}

func (u *authUseCase) Register(ctx context.Context, request dto_request.AuthRegisterRequest) {
	otp, err := u.repositoryManager.OtpRepository().GetByPhone(ctx, request.Phone)
	if err != nil {
		if err == constant.ErrNoData {
			panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuthOtpInvalid))
		}
		panic(err)
	}

	if otp.Verified || otp.ExpiresAt.IsLessThan(util.CurrentDateTime()) || otp.Otp != request.Otp {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuthOtpInvalid))
	}

	hashedPassword, hashErr := util.HashPassword(request.Password)
	panicIfErr(hashErr)

	birth, parseErr := time.Parse("2006-01-02", request.Birth)
	if parseErr != nil {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageSystemInvalidRequestPayload))
	}

	user := &model.User{
		Fullname: request.Fullname,
		Phone:    request.Phone,
		Birth:    data_type.NewDateTime(birth),
		Gender:   request.Gender,
		Password: hashedPassword,
	}

	if err := u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		if err := u.repositoryManager.UserRepository().Insert(ctx, user); err != nil {
			return err
		}
		return u.repositoryManager.OtpRepository().MarkVerified(ctx, request.Phone)
	}); err != nil {
		if err == constant.ErrDuplicateData {
			panic(dto_response.NewConflictErrorResponse(constant.LanguageAuthPhoneAlreadyRegistered))
		}
		panic(err)
	}
}

func (u *authUseCase) CreateOtp(ctx context.Context, phone string) {
	_, err := u.repositoryManager.UserRepository().GetByPhone(ctx, phone)
	if err == nil {
		// Phone already registered — silently succeed (don't reveal existence)
		return
	}
	if err != constant.ErrNoData {
		panic(err)
	}

	otpValue, genErr := util.GenerateOTP(6)
	panicIfErr(genErr)

	expiresAt := util.CurrentDateTime().Add(time.Minute)
	upsertErr := u.repositoryManager.OtpRepository().Upsert(ctx, phone, otpValue, expiresAt)
	panicIfErr(upsertErr)

	// TODO: send OTP via SMS/WhatsApp
	// For development, the OTP is accessible via GET /v1/auth/debug-otp (not exposed in production)
}

func (u *authUseCase) Parse(ctx context.Context, token string) (*model.UserClaims, error) {
	payload, err := u.jwt.Parse(token)
	if err != nil {
		return nil, constant.ErrNotAuthenticated
	}

	if util.CurrentDateTime().IsGreaterThan(payload.ExpiredAt) {
		return nil, constant.ErrNotAuthenticated
	}

	user, err := u.repositoryManager.UserRepository().GetById(ctx, payload.Id)
	if err != nil {
		if err == constant.ErrNoData {
			return nil, constant.ErrNotAuthenticated
		}
		return nil, err
	}

	userRoles, err := u.repositoryManager.UserRoleRepository().FetchByUserIds(ctx, []int64{payload.Id})
	if err != nil {
		return nil, err
	}

	roles := make([]string, 0, len(userRoles))
	for _, userRole := range userRoles {
		roles = append(roles, userRole.Role)
	}

	return &model.UserClaims{
		UserId: user.Id,
		Phone:  user.Phone,
		Roles:  roles,
	}, nil
}

func (u *authUseCase) SaveFcmToken(ctx context.Context, request dto_request.AuthFcmTokenRequest) {
	userClaims := model.MustGetUserCtx(ctx)
	// insert fcm token, take device info too
	userFcmToken := &model.UserFcmToken{
		UserId:   userClaims.UserId,
		FcmToken: request.FcmToken,
	}

	if err := u.repositoryManager.UserFcmTokenRepository().Insert(ctx, userFcmToken); err != nil {
		if err == constant.ErrDuplicateData {
			// Token already exists — silently succeed
			return
		}
		panic(err)
	}
}
