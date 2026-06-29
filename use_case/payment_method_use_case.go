package use_case

import (
	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/model"
	"auction-service/repository"
	"context"
)

type PaymentMethodUseCase interface {
	Create(ctx context.Context, req dto_request.PaymentMethodCreateRequest) model.PaymentMethod
	Fetch(ctx context.Context, req dto_request.PaymentMethodFetchRequest) ([]model.PaymentMethod, int64)
	AdminFetch(ctx context.Context, req dto_request.PaymentMethodFetchRequest) ([]model.PaymentMethod, int64)
	GetById(ctx context.Context, id int64) model.PaymentMethod
	Update(ctx context.Context, id int64, req dto_request.PaymentMethodUpdateRequest) model.PaymentMethod
}

type paymentMethodUseCase struct {
	repoManager repository.RepositoryManager
}

func NewPaymentMethodUseCase(rm repository.RepositoryManager) PaymentMethodUseCase {
	return &paymentMethodUseCase{repoManager: rm}
}

func (u *paymentMethodUseCase) Create(ctx context.Context, req dto_request.PaymentMethodCreateRequest) model.PaymentMethod {
	existing, _ := u.repoManager.PaymentMethodRepository().Fetch(ctx, model.PaymentMethodQueryOption{
		Code: &req.Code,
	})
	if len(existing) > 0 {
		panic(dto_response.NewConflictErrorResponse(constant.LanguagePaymentMethodCodeAlreadyUsed))
	}

	pm := model.PaymentMethod{
		Name:     req.Name,
		Code:     req.Code,
		Type:     req.Type,
		IsActive: true,
	}

	panicIfErr(u.repoManager.PaymentMethodRepository().Insert(ctx, &pm))

	return pm
}

func (u *paymentMethodUseCase) Fetch(ctx context.Context, req dto_request.PaymentMethodFetchRequest) ([]model.PaymentMethod, int64) {
	option := model.PaymentMethodQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(req.Page, req.Limit, nil),
		Name:        req.Name,
		Code:        req.Code,
		Type:        req.Type,
		IsActive:    req.IsActive,
	}

	paymentMethods, err := u.repoManager.PaymentMethodRepository().Fetch(ctx, option)
	panicIfErr(err)

	total, err := u.repoManager.PaymentMethodRepository().Count(ctx, option)
	panicIfErr(err)

	return paymentMethods, total
}

func (u *paymentMethodUseCase) GetById(ctx context.Context, id int64) model.PaymentMethod {
	pm, err := u.repoManager.PaymentMethodRepository().GetById(ctx, id)
	if err == constant.ErrNoData {
		panic(dto_response.NewNotFoundErrorResponse(constant.LanguagePaymentMethodNotFound))
	}
	panicIfErr(err)

	if pm == nil {
		panic(dto_response.NewNotFoundErrorResponse(constant.LanguagePaymentMethodNotFound))
	}

	return *pm
}

func (u *paymentMethodUseCase) AdminFetch(ctx context.Context, req dto_request.PaymentMethodFetchRequest) ([]model.PaymentMethod, int64) {
	return u.Fetch(ctx, req)
}

func (u *paymentMethodUseCase) Update(ctx context.Context, id int64, req dto_request.PaymentMethodUpdateRequest) model.PaymentMethod {
	pm, err := u.repoManager.PaymentMethodRepository().GetById(ctx, id)
	if err == constant.ErrNoData {
		panic(dto_response.NewNotFoundErrorResponse(constant.LanguagePaymentMethodNotFound))
	}
	panicIfErr(err)

	if pm == nil {
		panic(dto_response.NewNotFoundErrorResponse(constant.LanguagePaymentMethodNotFound))
	}

	if req.Name != nil {
		pm.Name = *req.Name
	}

	if req.Code != nil {
		pm.Code = *req.Code
	}

	if req.Type != nil {
		pm.Type = *req.Type
	}

	if req.IsActive != nil {
		pm.IsActive = *req.IsActive
	}

	pm, err = u.repoManager.PaymentMethodRepository().Update(ctx, pm)
	panicIfErr(err)

	return *pm
}
