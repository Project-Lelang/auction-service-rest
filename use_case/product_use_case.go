package use_case

import (
	"context"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/loader"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"

	"golang.org/x/sync/errgroup"
)

// ProductUseCase covers all product and product-status-history operations.
type ProductUseCase interface {
	// products
	OwnCreate(ctx context.Context, request dto_request.ProductCreateRequest) model.Product
	Fetch(ctx context.Context, request dto_request.ProductFetchRequest) ([]model.Product, int64)
	OwnFetch(ctx context.Context, request dto_request.OwnProductFetchRequest) ([]model.Product, int64)
	Get(ctx context.Context, productId string) model.Product
	AdminFetch(ctx context.Context, request dto_request.AdminProductFetchRequest) ([]model.Product, int64)

	// histories
	FetchStatusHistories(ctx context.Context, productId string) []model.ProductStatusHistory
	AdminFetchStatusHistories(ctx context.Context, productId string) []model.ProductStatusHistory
}

type productUseCase struct {
	repositoryManager repository.RepositoryManager
}

func NewProductUseCase(repositoryManager repository.RepositoryManager) ProductUseCase {
	return &productUseCase{repositoryManager: repositoryManager}
}

type productLoaderParams struct {
	user            bool
	statusHistories bool
}

func (u *productUseCase) mustLoadProductData(_ context.Context, products []*model.Product, option productLoaderParams) {
	userLoader := loader.NewUserLoader(u.repositoryManager.UserRepository())
	productStatusHistoriesLoader := loader.NewProductStatusHistoriesLoader(u.repositoryManager.ProductStatusHistoryRepository())

	panicIfErr(util.Await(func(group *errgroup.Group) {
		for _, product := range products {
			if option.user {
				group.Go(userLoader.ProductFn(product))
			}

			if option.statusHistories {
				group.Go(productStatusHistoriesLoader.ProductFn(product))
			}
		}
	}))
}

func (u *productUseCase) fetchPaginated(ctx context.Context, option model.ProductQueryOption) ([]model.Product, int64) {
	total, err := u.repositoryManager.ProductRepository().Count(ctx, option)
	panicIfErr(err)

	products, err := u.repositoryManager.ProductRepository().Fetch(ctx, option)
	panicIfErr(err)

	u.mustLoadProductData(ctx, util.SliceValueToSlicePointer(products), productLoaderParams{user: true})

	return products, total
}

func (u *productUseCase) OwnCreate(ctx context.Context, request dto_request.ProductCreateRequest) model.Product {
	userClaims := model.MustGetUserCtx(ctx)

	imageUrls := model.MarshalImageUrls(make([]string, 0, request.ImageCount))

	product := model.Product{
		Id:          util.NewUuid(),
		UserId:      userClaims.UserId,
		Name:        request.Name,
		Description: request.Description,
		Condition:   request.Condition,
		ImageUrls:   &imageUrls,
		Status:      constant.ProductStatusDraft,
	}

	panicIfErr(u.repositoryManager.ProductRepository().Insert(ctx, &product))

	return product
}

func (u *productUseCase) Fetch(ctx context.Context, request dto_request.ProductFetchRequest) ([]model.Product, int64) {
	return u.fetchPaginated(ctx, model.ProductQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(request.Page, request.Limit, model.Sorts(request.Sorts)),
		Status:      request.Status,
		Condition:   request.Condition,
		Search:      request.Search,
	})
}

func (u *productUseCase) OwnFetch(ctx context.Context, request dto_request.OwnProductFetchRequest) ([]model.Product, int64) {
	userClaims := model.MustGetUserCtx(ctx)
	return u.fetchPaginated(ctx, model.ProductQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(request.Page, request.Limit, model.Sorts(request.Sorts)),
		UserId:      &userClaims.UserId,
		Status:      request.Status,
		Condition:   request.Condition,
		Search:      request.Search,
	})
}

func (u *productUseCase) Get(ctx context.Context, productId string) model.Product {
	product := mustGetProduct(ctx, u.repositoryManager, productId)
	u.mustLoadProductData(ctx, []*model.Product{&product}, productLoaderParams{user: true, statusHistories: true})
	return product
}

func (u *productUseCase) AdminFetch(ctx context.Context, request dto_request.AdminProductFetchRequest) ([]model.Product, int64) {
	return u.fetchPaginated(ctx, model.ProductQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(request.Page, request.Limit, model.Sorts(request.Sorts)),
		Status:      request.Status,
		Condition:   request.Condition,
		Search:      request.Search,
	})
}

func (u *productUseCase) FetchStatusHistories(ctx context.Context, productId string) []model.ProductStatusHistory {
	mustGetProduct(ctx, u.repositoryManager, productId)

	histories, err := u.repositoryManager.ProductStatusHistoryRepository().FetchByProductId(ctx, productId)
	panicIfErr(err)

	return histories
}

func (u *productUseCase) AdminFetchStatusHistories(ctx context.Context, productId string) []model.ProductStatusHistory {
	mustGetProduct(ctx, u.repositoryManager, productId)

	histories, err := u.repositoryManager.ProductStatusHistoryRepository().FetchByProductId(ctx, productId)
	panicIfErr(err)

	return histories
}

func (u *productUseCase) recordHistory(ctx context.Context, productId string, status string, message *string) {
	h := &model.ProductStatusHistory{
		Id:        util.NewUuid(),
		ProductId: productId,
		Status:    status,
		Message:   message,
	}
	panicIfErr(u.repositoryManager.ProductStatusHistoryRepository().Insert(ctx, h))
}
