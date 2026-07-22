package use_case

import (
	"context"

	"auction-service/model"
	"auction-service/repository"
)

func updateProductStatusWithHistory(ctx context.Context, repositoryManager repository.RepositoryManager, productId int64, status string, message *string) error {
	if _, err := repositoryManager.ProductRepository().UpdateStatus(ctx, productId, status); err != nil {
		return err
	}
	return repositoryManager.ProductStatusHistoryRepository().Insert(ctx, &model.ProductStatusHistory{
		ProductId: productId,
		Status:    status,
		Message:   message,
	})
}
