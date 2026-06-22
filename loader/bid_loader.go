package loader

import (
	"context"

	"auction-service/model"
	"auction-service/repository"

	"github.com/graph-gophers/dataloader"
)

type BidLoader struct {
	loader dataloader.Loader
}

func (l *BidLoader) load(id int64) (*model.AuctionBid, error) {
	thunk := l.loader.Load(context.TODO(), int64Key(id))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.(*model.AuctionBid), nil
}

func (l *BidLoader) WinnerFn(winner *model.AuctionWinner) func() error {
	return func() error {
		if winner.AuctionBidId == nil {
			return nil
		}
		bid, err := l.load(*winner.AuctionBidId)
		if err != nil {
			return err
		}
		winner.AuctionBid = bid
		return nil
	}
}

func NewBidLoader(bidRepository repository.AuctionBidRepository) *BidLoader {
	batchFn := func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		ids := make([]int64, len(keys))
		for idx, k := range keys {
			ids[idx] = parseInt64Key(k)
		}

		bids, err := bidRepository.FetchByIds(ctx, ids)
		if err != nil {
			panic(err)
		}

		bidById := map[int64]model.AuctionBid{}
		for _, b := range bids {
			bidById[b.Id] = b
		}

		results := make([]*dataloader.Result, len(keys))
		for idx, k := range keys {
			var bid *model.AuctionBid
			if v, ok := bidById[parseInt64Key(k)]; ok {
				bid = &v
			}
			results[idx] = &dataloader.Result{Data: bid, Error: nil}
		}
		return results
	}

	return &BidLoader{
		loader: NewDataloader(batchFn),
	}
}
