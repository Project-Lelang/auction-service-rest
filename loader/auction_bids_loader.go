package loader

import (
	"context"

	"auction-service/model"
	"auction-service/repository"

	"github.com/graph-gophers/dataloader"
)

type AuctionBidsLoader struct {
	loader dataloader.Loader
}

func (l *AuctionBidsLoader) load(auctionId string) ([]*model.AuctionBid, error) {
	thunk := l.loader.Load(context.TODO(), dataloader.StringKey(auctionId))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.([]*model.AuctionBid), nil
}

func (l *AuctionBidsLoader) AuctionFn(auction *model.Auction) func() error {
	return func() error {
		bids, err := l.load(auction.Id)
		if err != nil {
			return err
		}
		auction.Bids = bids
		return nil
	}
}

func NewAuctionBidsLoader(bidRepository repository.AuctionBidRepository, userRepository repository.UserRepository) *AuctionBidsLoader {
	batchFn := func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		auctionIds := make([]string, len(keys))
		for idx, k := range keys {
			auctionIds[idx] = k.String()
		}

		bids, err := bidRepository.FetchByAuctionIds(ctx, auctionIds)
		if err != nil {
			panic(err)
		}

		// Load users for all bids
		userIds := make([]string, 0, len(bids))
		for _, bid := range bids {
			userIds = append(userIds, bid.UserId)
		}
		users, err := userRepository.FetchByIds(ctx, userIds)
		if err != nil {
			panic(err)
		}

		userById := map[string]model.User{}
		for _, u := range users {
			userById[u.Id] = u
		}

		// Attach users to bids
		for i := range bids {
			if user, ok := userById[bids[i].UserId]; ok {
				bids[i].User = &user
			}
		}

		// Group bids by auction ID
		bidsByAuctionId := map[string][]*model.AuctionBid{}
		for i := range bids {
			auctionId := bids[i].AuctionId
			bidsByAuctionId[auctionId] = append(bidsByAuctionId[auctionId], &bids[i])
		}

		results := make([]*dataloader.Result, len(keys))
		for idx, k := range keys {
			bids := bidsByAuctionId[k.String()]
			if bids == nil {
				bids = []*model.AuctionBid{}
			}
			results[idx] = &dataloader.Result{Data: bids, Error: nil}
		}
		return results
	}

	return &AuctionBidsLoader{
		loader: NewDataloader(batchFn),
	}
}
