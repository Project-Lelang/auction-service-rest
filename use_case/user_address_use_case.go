package use_case

import (
	"context"
	"strings"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/model"
	"auction-service/repository"
)

// UserAddressUseCase manages CRUD operations for user addresses.
type UserAddressUseCase interface {
	OwnCreate(ctx context.Context, request dto_request.UserAddressCreateRequest) model.UserAddress
	OwnFetch(ctx context.Context, request dto_request.UserAddressFetchRequest) ([]model.UserAddress, int)
	Get(ctx context.Context, request dto_request.UserAddressGetRequest) model.UserAddress
	OwnUpdate(ctx context.Context, request dto_request.UserAddressUpdateRequest) model.UserAddress
	OwnDelete(ctx context.Context, request dto_request.UserAddressDeleteRequest)
}

type userAddressUseCase struct {
	repositoryManager repository.RepositoryManager
}

func NewUserAddressUseCase(repositoryManager repository.RepositoryManager) UserAddressUseCase {
	return &userAddressUseCase{repositoryManager: repositoryManager}
}

func validateIndonesiaAddress(areaId string, latitude, longitude *float64) {
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(areaId)), "IDN") {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageUserAddressIndonesia))
	}
	if latitude == nil || longitude == nil {
		return
	}

	// Approximate Indonesia bounding box: western/eastern longitudes and
	// southern/northern latitudes across the archipelago.
	if *latitude < -11.0 || *latitude > 6.5 || *longitude < 95.0 || *longitude > 141.5 {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageUserAddressIndonesia))
	}
}

func (u *userAddressUseCase) mustGetOwned(ctx context.Context, id int64) model.UserAddress {
	userClaims := model.MustGetUserCtx(ctx)
	address, err := u.repositoryManager.UserAddressRepository().GetById(ctx, id)
	panicIfRepositoryError(err, constant.LanguageUserAddressNotFound)
	if address.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageUserAddressNotOwned))
	}
	return *address
}

func (u *userAddressUseCase) OwnCreate(ctx context.Context, request dto_request.UserAddressCreateRequest) model.UserAddress {
	userClaims := model.MustGetUserCtx(ctx)
	validateIndonesiaAddress(request.BiteshipAreaId, request.Latitude, request.Longitude)

	var err error
	if request.IsDefault {
		// Unset any existing default for this user
		err = u.repositoryManager.UserAddressRepository().UnsetDefaultByUserId(ctx, userClaims.UserId)
		panicIfErr(err)
	}

	address := model.UserAddress{
		UserId:         userClaims.UserId,
		Label:          request.Label,
		RecipientName:  request.RecipientName,
		Phone:          request.Phone,
		CityId:         request.CityId,
		CityName:       request.CityName,
		ProvinceName:   request.ProvinceName,
		Address:        request.Address,
		PostalCode:     request.PostalCode,
		BiteshipAreaId: request.BiteshipAreaId,
		Latitude:       request.Latitude,
		Longitude:      request.Longitude,
		IsDefault:      request.IsDefault,
	}
	panicIfErr(u.repositoryManager.UserAddressRepository().Insert(ctx, &address))
	return address
}

func (u *userAddressUseCase) OwnFetch(ctx context.Context, request dto_request.UserAddressFetchRequest) ([]model.UserAddress, int) {
	userClaims := model.MustGetUserCtx(ctx)

	option := model.UserAddressQueryOption{
		QueryOption: model.QueryOption{
			Page:  request.Page,
			Limit: request.Limit,
			Sorts: model.Sorts(request.Sorts),
		},
		UserId: &userClaims.UserId,
	}

	addresses, err := u.repositoryManager.UserAddressRepository().Fetch(ctx, option)
	panicIfErr(err)

	total, err := u.repositoryManager.UserAddressRepository().Count(ctx, option)
	panicIfErr(err)

	return addresses, total
}

func (u *userAddressUseCase) Get(ctx context.Context, request dto_request.UserAddressGetRequest) model.UserAddress {
	return u.mustGetOwned(ctx, request.UserAddressId)
}

func (u *userAddressUseCase) OwnUpdate(ctx context.Context, request dto_request.UserAddressUpdateRequest) model.UserAddress {
	address := u.mustGetOwned(ctx, request.UserAddressId)
	userClaims := model.MustGetUserCtx(ctx)
	validateIndonesiaAddress(request.BiteshipAreaId, request.Latitude, request.Longitude)

	if request.IsDefault && !address.IsDefault {
		panicIfErr(u.repositoryManager.UserAddressRepository().UnsetDefaultByUserId(ctx, userClaims.UserId))
	}

	address.Label = request.Label
	address.RecipientName = request.RecipientName
	address.Phone = request.Phone
	address.CityId = request.CityId
	address.CityName = request.CityName
	address.ProvinceName = request.ProvinceName
	address.Address = request.Address
	address.PostalCode = request.PostalCode
	address.BiteshipAreaId = request.BiteshipAreaId
	address.Latitude = request.Latitude
	address.Longitude = request.Longitude
	address.IsDefault = request.IsDefault

	updated, err := u.repositoryManager.UserAddressRepository().Update(ctx, &address)
	panicIfErr(err)
	return *updated
}

func (u *userAddressUseCase) OwnDelete(ctx context.Context, request dto_request.UserAddressDeleteRequest) {
	address := u.mustGetOwned(ctx, request.UserAddressId)
	if address.IsDefault {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageUserAddressIsDefault))
	}
	panicIfErr(u.repositoryManager.UserAddressRepository().Delete(ctx, address.Id))
}
