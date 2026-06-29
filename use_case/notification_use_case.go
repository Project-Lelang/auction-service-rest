package use_case

import (
	"context"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/model"
	"auction-service/repository"
)

type NotificationUseCase interface {
	OwnFetch(ctx context.Context, request dto_request.OwnNotificationFetchRequest) ([]model.Notification, int64)
	OwnGet(ctx context.Context, request dto_request.OwnNotificationGetRequest) model.Notification
	OwnMarkRead(ctx context.Context, request dto_request.OwnNotificationMarkReadRequest) model.Notification
}

type notificationUseCase struct {
	repositoryManager repository.RepositoryManager
}

func NewNotificationUseCase(repositoryManager repository.RepositoryManager) NotificationUseCase {
	return &notificationUseCase{repositoryManager: repositoryManager}
}

func (u *notificationUseCase) OwnFetch(ctx context.Context, request dto_request.OwnNotificationFetchRequest) ([]model.Notification, int64) {
	userClaims := model.MustGetUserCtx(ctx)

	option := model.NotificationQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(
			request.Page,
			request.Limit,
			model.Sorts(request.Sorts),
		),
		UserId: &userClaims.UserId,
		IsRead: request.IsRead,
	}

	total, err := u.repositoryManager.NotifcationRepository().Count(ctx, option)
	panicIfErr(err)

	notifications, err := u.repositoryManager.NotifcationRepository().Fetch(ctx, option)
	panicIfErr(err)

	return notifications, total
}

func (u *notificationUseCase) OwnGet(ctx context.Context, request dto_request.OwnNotificationGetRequest) model.Notification {
	return u.mustGetOwned(ctx, request.NotificationId)
}

func (u *notificationUseCase) OwnMarkRead(ctx context.Context, request dto_request.OwnNotificationMarkReadRequest) model.Notification {
	notification := u.mustGetOwned(ctx, request.NotificationId)

	updated, err := u.repositoryManager.NotifcationRepository().MarkRead(ctx, notification.Id)
	panicIfErr(err)

	return *updated
}

func (u *notificationUseCase) mustGetOwned(ctx context.Context, notificationId int64) model.Notification {
	userClaims := model.MustGetUserCtx(ctx)

	notification, err := u.repositoryManager.NotifcationRepository().GetById(ctx, notificationId)
	if err == constant.ErrNoData {
		panic(dto_response.NewNotFoundErrorResponse(constant.LanguageNotificationNotFound))
	}
	panicIfErr(err)

	if notification.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	return *notification
}
