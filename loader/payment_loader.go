package loader

import (
	"context"

	"auction-service/model"
	"auction-service/repository"

	"github.com/graph-gophers/dataloader"
)

type PaymentLoader struct {
	loader dataloader.Loader
}

func (l *PaymentLoader) load(auctionId int64) (*model.Payment, error) {
	thunk := l.loader.Load(context.TODO(), int64Key(auctionId))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.(*model.Payment), nil
}

func (l *PaymentLoader) AuctionFn(auction *model.Auction) func() error {
	return func() error {
		payment, err := l.load(auction.Id)
		if err != nil {
			return err
		}
		auction.Payment = payment
		return nil
	}
}

func NewPaymentLoader(paymentRepository repository.PaymentRepository) *PaymentLoader {
	batchFn := func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		auctionIds := make([]int64, len(keys))
		for idx, k := range keys {
			auctionIds[idx] = parseInt64Key(k)
		}

		payments, err := paymentRepository.FetchByAuctionIds(ctx, auctionIds)
		if err != nil {
			panic(err)
		}

		paymentByAuctionId := map[int64]model.Payment{}
		for _, p := range payments {
			paymentByAuctionId[p.AuctionId] = p
		}

		results := make([]*dataloader.Result, len(keys))
		for idx, k := range keys {
			var payment *model.Payment
			if v, ok := paymentByAuctionId[parseInt64Key(k)]; ok {
				payment = &v
			}
			results[idx] = &dataloader.Result{Data: payment, Error: nil}
		}
		return results
	}

	return &PaymentLoader{
		loader: NewDataloader(batchFn),
	}
}
