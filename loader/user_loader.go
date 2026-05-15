package loader

import (
	"context"

	"auction-service/constant"
	"auction-service/model"
	"auction-service/repository"

	"github.com/graph-gophers/dataloader"
)

type UserLoader struct {
	loader dataloader.Loader
}

func (l *UserLoader) load(id string) (*model.User, error) {
	thunk := l.loader.Load(context.TODO(), dataloader.StringKey(id))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.(*model.User), nil
}

func (l *UserLoader) ProductFn(product *model.Product) func() error {
	return func() error {
		user, err := l.load(product.UserId)
		if err != nil {
			return err
		}
		product.User = user
		return nil
	}
}

func NewUserLoader(userRepository repository.UserRepository) *UserLoader {
	batchFn := func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		ids := make([]string, len(keys))
		for idx, k := range keys {
			ids[idx] = k.String()
		}

		users, err := userRepository.FetchByIds(ctx, ids)
		if err != nil {
			panic(err)
		}

		userById := map[string]model.User{}
		for _, user := range users {
			userById[user.Id] = user
		}

		results := make([]*dataloader.Result, len(keys))
		for idx, k := range keys {
			var user *model.User
			if v, ok := userById[k.String()]; ok {
				user = &v
			}
			result := &dataloader.Result{Data: user, Error: nil}
			if user == nil {
				result.Error = constant.ErrNoData
			}
			results[idx] = result
		}
		return results
	}

	return &UserLoader{
		loader: NewDataloader(batchFn),
	}
}
