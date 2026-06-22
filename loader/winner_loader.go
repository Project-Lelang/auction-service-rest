package loader

import (
	"context"

	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"

	"github.com/graph-gophers/dataloader"
	"golang.org/x/sync/errgroup"
)

type WinnerLoader struct {
	loader         dataloader.Loader
	bidRepository  repository.AuctionBidRepository
	userRepository repository.UserRepository
}

func (l *WinnerLoader) load(auctionId int64) (*model.AuctionWinner, error) {
	thunk := l.loader.Load(context.TODO(), int64Key(auctionId))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.(*model.AuctionWinner), nil
}

func (l *WinnerLoader) AuctionFn(auction *model.Auction) func() error {
	return func() error {
		winner, err := l.load(auction.Id)
		if err != nil {
			return err
		}
		if winner != nil {
			// Load bid and user for the winner
			bidLoader := NewBidLoader(l.bidRepository)
			userLoader := NewUserLoader(l.userRepository)

			_ = util.Await(func(group *errgroup.Group) {
				group.Go(bidLoader.WinnerFn(winner))
			})

			if winner.AuctionBid != nil {
				_ = util.Await(func(group *errgroup.Group) {
					group.Go(func() error {
						user, err := userLoader.load(winner.AuctionBid.UserId)
						if err != nil {
							return err
						}
						winner.AuctionBid.User = user
						return nil
					})
				})
			}
		}
		auction.Winner = winner
		return nil
	}
}

func NewWinnerLoader(winnerRepository repository.AuctionWinnerRepository, bidRepository repository.AuctionBidRepository, userRepository repository.UserRepository) *WinnerLoader {
	batchFn := func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		auctionIds := make([]int64, len(keys))
		for idx, k := range keys {
			auctionIds[idx] = parseInt64Key(k)
		}

		winners, err := winnerRepository.FetchByAuctionIds(ctx, auctionIds)
		if err != nil {
			panic(err)
		}

		winnerByAuctionId := map[int64]model.AuctionWinner{}
		for _, w := range winners {
			winnerByAuctionId[w.AuctionId] = w
		}

		results := make([]*dataloader.Result, len(keys))
		for idx, k := range keys {
			var winner *model.AuctionWinner
			if v, ok := winnerByAuctionId[parseInt64Key(k)]; ok {
				winner = &v
			}
			results[idx] = &dataloader.Result{Data: winner, Error: nil}
		}
		return results
	}

	return &WinnerLoader{
		loader:         NewDataloader(batchFn),
		bidRepository:  bidRepository,
		userRepository: userRepository,
	}
}
