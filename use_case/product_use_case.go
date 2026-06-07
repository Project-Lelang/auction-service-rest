package use_case

import (
	"context"
	"fmt"
	"path"
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

// ProductUseCase covers all product and product-status-history operations.
type ProductUseCase interface {
	// products
	OwnCreate(ctx context.Context, request dto_request.OwnProductCreateRequest) model.Product
	OwnGet(ctx context.Context, request dto_request.OwnProductGetRequest) model.Product
	OwnUpdate(ctx context.Context, request dto_request.OwnProductUpdateRequest) model.Product
	OwnRequest(ctx context.Context, request dto_request.OwnProductRequestRequest) model.Product
	Fetch(ctx context.Context, request dto_request.ProductFetchRequest) ([]model.Product, int64)
	OwnFetch(ctx context.Context, request dto_request.OwnProductFetchRequest) ([]model.Product, int64)
	Get(ctx context.Context, request dto_request.ProductGetRequest) model.Product
	AdminFetch(ctx context.Context, request dto_request.AdminProductFetchRequest) ([]model.Product, int64)
	AdminApprove(ctx context.Context, request dto_request.AdminProductApproveRequest) model.Product
	AdminReject(ctx context.Context, request dto_request.AdminProductRejectRequest) model.Product

	// histories
	OwnFetchStatusHistories(ctx context.Context, request dto_request.OwnProductFetchStatusHistoriesRequest) []model.ProductStatusHistory
	AdminFetchStatusHistories(ctx context.Context, request dto_request.AdminProductFetchStatusHistoriesRequest) []model.ProductStatusHistory
}

type productUseCase struct {
	BaseFileUseCase
	repositoryManager repository.RepositoryManager
	filesystemManager internalFilesystem.FilesystemManager
}

func NewProductUseCase(repositoryManager repository.RepositoryManager, filesystemManager internalFilesystem.FilesystemManager) ProductUseCase {
	return &productUseCase{
		BaseFileUseCase:   NewBaseFileUseCase(filesystemManager.Main(), filesystemManager.Tmp()),
		repositoryManager: repositoryManager,
		filesystemManager: filesystemManager,
	}
}

func (u *productUseCase) populateImageLinks(products ...*model.Product) {
	const presignedExpiry = 24 * time.Hour
	mainFs := u.filesystemManager.Main()
	for _, product := range products {
		if product.CoverImagePath != nil && *product.CoverImagePath != "" {
			link := mainFs.PresignedUrl(util.GetFilenameFromPath(*product.CoverImagePath), *product.CoverImagePath, presignedExpiry)
			product.CoverImageLink = &link
		}
		imagePaths := model.ParseImagePaths(product.ImagePaths)
		links := make([]string, 0, len(imagePaths))
		for _, p := range imagePaths {
			links = append(links, mainFs.PresignedUrl(util.GetFilenameFromPath(p), p, presignedExpiry))
		}
		product.ImageLinks = links
	}
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
	u.populateImageLinks(util.SliceValueToSlicePointer(products)...)

	return products, total
}

func (u *productUseCase) OwnCreate(ctx context.Context, request dto_request.OwnProductCreateRequest) model.Product {
	userClaims := model.MustGetUserCtx(ctx)
	productId := util.NewUuid()

	// Validate all tmp paths upfront for a fast fail before any file I/O.
	allTmpPaths := []string{}
	if request.CoverImagePath != nil {
		allTmpPaths = append(allTmpPaths, *request.CoverImagePath)
	}
	allTmpPaths = append(allTmpPaths, request.ImagePaths...)
	if len(allTmpPaths) > 0 {
		u.MustValidateTemporaryFilePaths(allTmpPaths)
	}

	// Move cover image from tmp to main storage.
	var coverMainPath *string
	if request.CoverImagePath != nil {
		p, _ := u.MustUploadFileFromTemporaryToMain(ctx, "products", productId,
			"cover"+path.Ext(*request.CoverImagePath), *request.CoverImagePath,
			FileUploadTemporaryToMainParams{DeleteTmpOnSuccess: true})
		coverMainPath = &p
	}

	// Move additional images from tmp to main storage.
	mainImagePaths := make([]string, 0, len(request.ImagePaths))
	for i, tmpPath := range request.ImagePaths {
		p, _ := u.MustUploadFileFromTemporaryToMain(ctx, "products", productId,
			fmt.Sprintf("image_%d%s", i, path.Ext(tmpPath)), tmpPath,
			FileUploadTemporaryToMainParams{DeleteTmpOnSuccess: true})
		mainImagePaths = append(mainImagePaths, p)
	}

	imagePathsStr := model.MarshalImagePaths(mainImagePaths)
	product := model.Product{
		Id:             productId,
		UserId:         userClaims.UserId,
		Name:           request.Name,
		Description:    request.Description,
		Condition:      request.Condition,
		CoverImagePath: coverMainPath,
		ImagePaths:     &imagePathsStr,
		WeightGram:     request.WeightGram,
		Status:         constant.ProductStatusDraft,
	}

	panicIfErr(u.repositoryManager.ProductRepository().Insert(ctx, &product))

	u.populateImageLinks(&product)
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

func (u *productUseCase) Get(ctx context.Context, request dto_request.ProductGetRequest) model.Product {
	product := mustGetProduct(ctx, u.repositoryManager, request.ProductId)
	u.mustLoadProductData(ctx, []*model.Product{&product}, productLoaderParams{user: true, statusHistories: true})
	u.populateImageLinks(&product)
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

func (u *productUseCase) AdminApprove(ctx context.Context, request dto_request.AdminProductApproveRequest) model.Product {
	product := mustGetProduct(ctx, u.repositoryManager, request.ProductId)

	// Validate ownership alignment
	if product.UserId != request.UserId {
		panic(dto_response.NewBadRequestErrorResponse("Product does not belong to specified user"))
	}
	// Verify it can transition from REQUEST to VERIFIED status
	if product.Status != constant.ProductStatusRequest {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageProductInvalidStatusTransition))
	}

	var updated *model.Product
	panicIfErr(u.repositoryManager.Transaction(ctx, func(txCtx context.Context) error {
		var err error
		updated, err = u.repositoryManager.ProductRepository().UpdateStatus(txCtx, request.ProductId, constant.ProductStatusVerified)
		if err != nil {
			return err
		}

		return u.repositoryManager.ProductStatusHistoryRepository().Insert(txCtx, &model.ProductStatusHistory{
			Id:        util.NewUuid(),
			ProductId: request.ProductId,
			Status:    constant.ProductStatusVerified,
			Message:   nil, // Verification holds no negative comment message
		})
	}))

	u.populateImageLinks(updated)
	return *updated
}

func (u *productUseCase) AdminReject(ctx context.Context, request dto_request.AdminProductRejectRequest) model.Product {
	product := mustGetProduct(ctx, u.repositoryManager, request.ProductId)

	if product.UserId != request.UserId {
		panic(dto_response.NewBadRequestErrorResponse("Product does not belong to specified user"))
	}
	if product.Status != constant.ProductStatusRequest {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageProductInvalidStatusTransition))
	}

	var updated *model.Product
	panicIfErr(u.repositoryManager.Transaction(ctx, func(txCtx context.Context) error {
		var err error
		updated, err = u.repositoryManager.ProductRepository().UpdateStatus(txCtx, request.ProductId, constant.ProductStatusRejected)
		if err != nil {
			return err
		}

		return u.repositoryManager.ProductStatusHistoryRepository().Insert(txCtx, &model.ProductStatusHistory{
			Id:        util.NewUuid(),
			ProductId: request.ProductId,
			Status:    constant.ProductStatusRejected,
			Message:   request.Message, // Saved directly to history record
		})
	}))

	u.populateImageLinks(updated)
	return *updated
}

func (u *productUseCase) OwnFetchStatusHistories(ctx context.Context, request dto_request.OwnProductFetchStatusHistoriesRequest) []model.ProductStatusHistory {
	mustGetProduct(ctx, u.repositoryManager, request.ProductId)

	histories, err := u.repositoryManager.ProductStatusHistoryRepository().FetchByProductId(ctx, request.ProductId)
	panicIfErr(err)

	return histories
}

func (u *productUseCase) AdminFetchStatusHistories(ctx context.Context, request dto_request.AdminProductFetchStatusHistoriesRequest) []model.ProductStatusHistory {
	mustGetProduct(ctx, u.repositoryManager, request.ProductId)

	histories, err := u.repositoryManager.ProductStatusHistoryRepository().FetchByProductId(ctx, request.ProductId)
	panicIfErr(err)

	return histories
}

func (u *productUseCase) OwnGet(ctx context.Context, request dto_request.OwnProductGetRequest) model.Product {
	userClaims := model.MustGetUserCtx(ctx)

	product := mustGetProduct(ctx, u.repositoryManager, request.ProductId)
	if product.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	u.mustLoadProductData(ctx, []*model.Product{&product}, productLoaderParams{user: true, statusHistories: true})
	u.populateImageLinks(&product)
	return product
}

func (u *productUseCase) OwnUpdate(ctx context.Context, request dto_request.OwnProductUpdateRequest) model.Product {
	userClaims := model.MustGetUserCtx(ctx)

	product := mustGetProduct(ctx, u.repositoryManager, request.ProductId)
	if product.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}
	if product.Status != constant.ProductStatusDraft {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageProductInvalidStatusTransition))
	}

	// Validate all new tmp paths upfront before doing any file I/O.
	allTmpPaths := []string{}
	if request.CoverImagePath != nil {
		allTmpPaths = append(allTmpPaths, *request.CoverImagePath)
	}
	allTmpPaths = append(allTmpPaths, request.ImagePaths...)
	if len(allTmpPaths) > 0 {
		u.MustValidateTemporaryFilePaths(allTmpPaths)
	}

	// Update text fields.
	updated, err := u.repositoryManager.ProductRepository().Update(ctx, request.ProductId, request.Name, request.Description, request.Condition, request.WeightGram)
	panicIfErr(err)

	// Handle image updates only when the caller explicitly provides new paths.
	if request.CoverImagePath != nil || request.ImagePaths != nil {
		newCoverPath := updated.CoverImagePath // default: keep existing

		if request.CoverImagePath != nil {
			// Delete old cover asynchronously.
			if product.CoverImagePath != nil && *product.CoverImagePath != "" {
				go u.filesystemManager.Main().Delete(*product.CoverImagePath) //nolint:errcheck
			}
			// Upload new cover.
			p, _ := u.MustUploadFileFromTemporaryToMain(ctx, "products", request.ProductId,
				"cover"+path.Ext(*request.CoverImagePath), *request.CoverImagePath,
				FileUploadTemporaryToMainParams{DeleteTmpOnSuccess: true})
			newCoverPath = &p
		}

		newImagePathsStr := updated.ImagePaths // default: keep existing

		if request.ImagePaths != nil {
			// Delete all old images asynchronously.
			for _, oldPath := range model.ParseImagePaths(product.ImagePaths) {
				oldPath := oldPath
				go u.filesystemManager.Main().Delete(oldPath) //nolint:errcheck
			}
			// Upload new images.
			mainImagePaths := make([]string, 0, len(request.ImagePaths))
			for i, tmpPath := range request.ImagePaths {
				p, _ := u.MustUploadFileFromTemporaryToMain(ctx, "products", request.ProductId,
					fmt.Sprintf("image_%d%s", i, path.Ext(tmpPath)), tmpPath,
					FileUploadTemporaryToMainParams{DeleteTmpOnSuccess: true})
				mainImagePaths = append(mainImagePaths, p)
			}
			s := model.MarshalImagePaths(mainImagePaths)
			newImagePathsStr = &s
		}

		updated, err = u.repositoryManager.ProductRepository().UpdateImages(ctx, request.ProductId, newCoverPath, newImagePathsStr)
		panicIfErr(err)
	}

	u.populateImageLinks(updated)
	return *updated
}

func (u *productUseCase) OwnRequest(ctx context.Context, request dto_request.OwnProductRequestRequest) model.Product {
	userClaims := model.MustGetUserCtx(ctx)

	product := mustGetProduct(ctx, u.repositoryManager, request.ProductId)
	if product.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}
	if !constant.ValidProductStatusTransitionFor(product.Status, constant.ProductStatusRequest) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageProductInvalidStatusTransition))
	}

	var updated *model.Product
	panicIfErr(u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		var err error
		updated, err = u.repositoryManager.ProductRepository().UpdateStatus(ctx, request.ProductId, constant.ProductStatusRequest)
		if err != nil {
			return err
		}
		return u.repositoryManager.ProductStatusHistoryRepository().Insert(ctx, &model.ProductStatusHistory{
			Id:        util.NewUuid(),
			ProductId: request.ProductId,
			Status:    constant.ProductStatusRequest,
		})
	}))

	u.mustLoadProductData(ctx, []*model.Product{updated}, productLoaderParams{statusHistories: true})
	u.populateImageLinks(updated)
	return *updated
}
