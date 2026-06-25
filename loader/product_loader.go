package loader

import (
	"context"

	"auction-service/model"
	"auction-service/repository"

	"github.com/graph-gophers/dataloader"
)

type ProductLoader struct {
	loader dataloader.Loader
}

func (l *ProductLoader) load(id int64) (*model.Product, error) {
	thunk := l.loader.Load(context.TODO(), int64Key(id))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.(*model.Product), nil
}

func (l *ProductLoader) AuctionFn(auction *model.Auction) func() error {
	return func() error {
		product, err := l.load(auction.ProductId)
		if err != nil {
			return err
		}
		auction.Product = product
		return nil
	}
}

func NewProductLoader(productRepository repository.ProductRepository) *ProductLoader {
	batchFn := func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		ids := make([]int64, len(keys))
		for idx, k := range keys {
			ids[idx] = parseInt64Key(k)
		}

		products, err := productRepository.FetchByIds(ctx, ids)
		if err != nil {
			panic(err)
		}

		productById := map[int64]model.Product{}
		for _, p := range products {
			productById[p.Id] = p
		}

		results := make([]*dataloader.Result, len(keys))
		for idx, k := range keys {
			var product *model.Product
			if v, ok := productById[parseInt64Key(k)]; ok {
				product = &v
			}
			results[idx] = &dataloader.Result{Data: product, Error: nil}
		}
		return results
	}

	return &ProductLoader{
		loader: NewDataloader(batchFn),
	}
}

type UserRolesLoader struct {
	loader dataloader.Loader
}

func (l *UserRolesLoader) load(userId int64) ([]model.UserRole, error) {
	thunk := l.loader.Load(context.TODO(), int64Key(userId))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.([]model.UserRole), nil
}

func (l *UserRolesLoader) UserFn(user *model.User) func() error {
	return func() error {
		roles, err := l.load(user.Id)
		if err != nil {
			return err
		}
		user.Roles = roles
		return nil
	}
}

func NewUserRolesLoader(userRoleRepository repository.UserRoleRepository) *UserRolesLoader {
	batchFn := func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		userIds := make([]int64, len(keys))
		for idx, k := range keys {
			userIds[idx] = parseInt64Key(k)
		}

		roles, err := userRoleRepository.FetchByUserIds(ctx, userIds)
		if err != nil {
			panic(err)
		}

		rolesByUserId := map[int64][]model.UserRole{}
		for _, r := range roles {
			rolesByUserId[r.UserId] = append(rolesByUserId[r.UserId], r)
		}

		results := make([]*dataloader.Result, len(keys))
		for idx, k := range keys {
			userRoles := []model.UserRole{}
			if v, ok := rolesByUserId[parseInt64Key(k)]; ok {
				userRoles = v
			}
			results[idx] = &dataloader.Result{Data: userRoles, Error: nil}
		}
		return results
	}

	return &UserRolesLoader{
		loader: NewDataloader(batchFn),
	}
}

type ProductStatusHistoriesLoader struct {
	loader dataloader.Loader
}

func (l *ProductStatusHistoriesLoader) load(productId int64) ([]model.ProductStatusHistory, error) {
	thunk := l.loader.Load(context.TODO(), int64Key(productId))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.([]model.ProductStatusHistory), nil
}

func (l *ProductStatusHistoriesLoader) ProductFn(product *model.Product) func() error {
	return func() error {
		histories, err := l.load(product.Id)
		if err != nil {
			return err
		}
		product.StatusHistories = histories
		return nil
	}
}

func NewProductStatusHistoriesLoader(repo repository.ProductStatusHistoryRepository) *ProductStatusHistoriesLoader {
	batchFn := func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		productIds := make([]int64, len(keys))
		for idx, k := range keys {
			productIds[idx] = parseInt64Key(k)
		}

		histories, err := repo.FetchByProductIds(ctx, productIds)
		if err != nil {
			panic(err)
		}

		historiesByProductId := map[int64][]model.ProductStatusHistory{}
		for _, h := range histories {
			historiesByProductId[h.ProductId] = append(historiesByProductId[h.ProductId], h)
		}

		results := make([]*dataloader.Result, len(keys))
		for idx, k := range keys {
			productHistories := []model.ProductStatusHistory{}
			if v, ok := historiesByProductId[parseInt64Key(k)]; ok {
				productHistories = v
			}
			results[idx] = &dataloader.Result{Data: productHistories, Error: nil}
		}
		return results
	}

	return &ProductStatusHistoriesLoader{
		loader: NewDataloader(batchFn),
	}
}
