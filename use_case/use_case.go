package use_case

import (
	"context"

	"auction-service/constant"
	"auction-service/delivery/dto_response"
	"auction-service/model"
	"auction-service/repository"
)

func panicIfErr(err error, excludedErrs ...error) {
	if err != nil {
		for _, excludedErr := range excludedErrs {
			if err == excludedErr {
				return
			}
		}
		panic(err)
	}
}

func panicIfRepositoryError(err error, errNoDataMessage string) {
	if err != nil {
		if err == constant.ErrNoData {
			panic(dto_response.NewBadRequestErrorResponse(errNoDataMessage))
		}
		panic(err)
	}
}

func mustGetUser(ctx context.Context, repositoryManager repository.RepositoryManager, id string) model.User {
	user, err := repositoryManager.UserRepository().GetById(ctx, id)
	panicIfRepositoryError(err, constant.LanguageUserNotFound)
	return *user
}

func mustGetProduct(ctx context.Context, repositoryManager repository.RepositoryManager, id string) model.Product {
	product, err := repositoryManager.ProductRepository().GetById(ctx, id)
	panicIfRepositoryError(err, constant.LanguageProductNotFound)
	return *product
}

func mustGetAuction(ctx context.Context, repositoryManager repository.RepositoryManager, id string) model.Auction {
	auction, err := repositoryManager.AuctionRepository().GetById(ctx, id)
	panicIfRepositoryError(err, constant.LanguageAuctionNotFound)
	return *auction
}

func mustGetAuctionWinner(ctx context.Context, repositoryManager repository.RepositoryManager, id string) model.AuctionWinner {
	winner, err := repositoryManager.AuctionWinnerRepository().GetById(ctx, id)
	panicIfRepositoryError(err, constant.LanguageWinnerNotFound)
	return *winner
}

func mustGetPayment(ctx context.Context, repositoryManager repository.RepositoryManager, id string) model.Payment {
	payment, err := repositoryManager.PaymentRepository().GetById(ctx, id)
	panicIfRepositoryError(err, constant.LanguagePaymentNotFound)
	return *payment
}

func mustGetShipment(ctx context.Context, repositoryManager repository.RepositoryManager, id string) model.Shipment {
	shipment, err := repositoryManager.ShipmentRepository().GetById(ctx, id)
	panicIfRepositoryError(err, constant.LanguageShipmentNotFound)
	return *shipment
}

func mustGetAuctionBid(ctx context.Context, repositoryManager repository.RepositoryManager, id string) model.AuctionBid {
	bid, err := repositoryManager.AuctionBidRepository().GetById(ctx, id)
	panicIfRepositoryError(err, constant.LanguageBidNotFound)
	return *bid
}
