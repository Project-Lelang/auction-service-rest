package use_case

import (
	"context"
	"time"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/global"
	internalJwt "auction-service/internal/jwt"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"
)

type AuthUseCase interface {
	Login(ctx context.Context, request dto_request.AuthLoginRequest) model.Token
	Refresh(ctx context.Context, request dto_request.AuthRefreshTokenRequest) model.Token
	Logout(ctx context.Context)
	Parse(ctx context.Context, token string) (*model.UserAccessToken, *model.User, error)
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

func (u *authUseCase) generateTokenPair(ctx context.Context, user *model.User) model.Token {
	now := time.Now().UTC()
	jwtConfig := global.GetConfig().JwtConfig

	accessTokenId := util.NewUuid()
	accessTokenExpiredAt := now.Add(time.Duration(jwtConfig.AccessTokenExpiryHours) * time.Hour)

	accessToken, err := u.jwt.Generate(internalJwt.Payload{
		UserAccessTokenId: accessTokenId,
		UserId:            user.Id,
		CreatedAt:         now,
		ExpiredAt:         accessTokenExpiredAt,
	})
	panicIfErr(err)

	// save access token
	userAccessToken := &model.UserAccessToken{
		Id:        accessTokenId,
		UserId:    user.Id,
		Token:     accessToken.AccessToken,
		ExpiredAt: accessTokenExpiredAt,
		CreatedAt: now,
	}
	err = u.repositoryManager.UserAccessTokenRepository().Insert(ctx, userAccessToken)
	panicIfErr(err)

	refreshTokenId := util.NewUuid()
	refreshTokenExpiredAt := now.Add(time.Duration(jwtConfig.RefreshTokenExpiryHours) * time.Hour)

	refreshToken, err := u.jwt.Generate(internalJwt.Payload{
		UserAccessTokenId: refreshTokenId,
		UserId:            user.Id,
		CreatedAt:         now,
		ExpiredAt:         refreshTokenExpiredAt,
	})
	panicIfErr(err)

	// save refresh token
	refreshUserAccessToken := &model.UserAccessToken{
		Id:        refreshTokenId,
		UserId:    user.Id,
		Token:     refreshToken.AccessToken,
		ExpiredAt: refreshTokenExpiredAt,
		CreatedAt: now,
	}
	err = u.repositoryManager.UserAccessTokenRepository().Insert(ctx, refreshUserAccessToken)
	panicIfErr(err)

	return model.Token{
		AccessToken:           "Bearer " + accessToken.AccessToken,
		AccessTokenExpiredAt:  accessTokenExpiredAt,
		RefreshToken:          "Bearer " + refreshToken.AccessToken,
		RefreshTokenExpiredAt: refreshTokenExpiredAt,
		TokenType:             "Bearer",
	}
}

func (u *authUseCase) Login(ctx context.Context, request dto_request.AuthLoginRequest) model.Token {
	user, err := u.repositoryManager.UserRepository().GetByEmail(ctx, request.Email)
	if err != nil {
		if err == constant.ErrNoData {
			panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuthInvalidCredentials))
		}
		panic(err)
	}

	if !util.CheckPasswordHash(request.Password, user.Password) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuthInvalidCredentials))
	}

	return u.generateTokenPair(ctx, user)
}

func (u *authUseCase) Refresh(ctx context.Context, request dto_request.AuthRefreshTokenRequest) model.Token {
	payload, err := u.jwt.Parse(request.RefreshToken)
	if err != nil {
		panic(dto_response.NewUnauthorizedErrorResponse(constant.LanguageAuthRefreshTokenInvalid))
	}

	if time.Now().UTC().After(payload.ExpiredAt) {
		panic(dto_response.NewUnauthorizedErrorResponse(constant.LanguageAuthRefreshTokenExpired))
	}

	// validate token exists in DB
	tokenRecord, err := u.repositoryManager.UserAccessTokenRepository().GetById(ctx, payload.UserAccessTokenId)
	if err != nil {
		panic(dto_response.NewUnauthorizedErrorResponse(constant.LanguageAuthRefreshTokenInvalid))
	}

	user, err := u.repositoryManager.UserRepository().Get(ctx, tokenRecord.UserId)
	panicIfRepositoryError(err, constant.LanguageUserNotFound)

	// delete old token pair
	err = u.repositoryManager.UserAccessTokenRepository().DeleteByUserId(ctx, user.Id)
	panicIfErr(err)

	return u.generateTokenPair(ctx, user)
}

func (u *authUseCase) Logout(ctx context.Context) {
	userAccessToken, err := model.GetUserAccessTokenCtx(ctx)
	if err != nil {
		panic(constant.ErrNotAuthenticated)
	}

	err = u.repositoryManager.UserAccessTokenRepository().Delete(ctx, userAccessToken)
	panicIfErr(err)
}

func (u *authUseCase) Parse(ctx context.Context, token string) (*model.UserAccessToken, *model.User, error) {
	payload, err := u.jwt.Parse(token)
	if err != nil {
		return nil, nil, constant.ErrNotAuthenticated
	}

	if time.Now().UTC().After(payload.ExpiredAt) {
		return nil, nil, constant.ErrNotAuthenticated
	}

	tokenRecord, err := u.repositoryManager.UserAccessTokenRepository().GetById(ctx, payload.UserAccessTokenId)
	if err != nil {
		return nil, nil, constant.ErrNotAuthenticated
	}

	user, err := u.repositoryManager.UserRepository().Get(ctx, tokenRecord.UserId)
	if err != nil {
		return nil, nil, constant.ErrNotAuthenticated
	}

	return tokenRecord, user, nil
}
