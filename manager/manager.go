package manager

import (
	"auction-service/delivery/ws"
	"auction-service/global"
	"auction-service/infrastructure"
	internalFilesystem "auction-service/internal/filesystem"
	internalJwt "auction-service/internal/jwt"
	"auction-service/repository"
	"auction-service/use_case"
)

type Container struct {
	infrastructureManager infrastructure.InfrastructureManager
	repositoryManager     repository.RepositoryManager
	useCaseManager        use_case.UseCaseManager
	filesystemManager     internalFilesystem.FilesystemManager
	baseFileUseCase       use_case.BaseFileUseCase
}

func NewContainer(hub *ws.Hub) *Container {
	config := global.GetConfig()

	infraManager := infrastructure.NewInfrastructureManager(config)

	db := infraManager.GetDB()

	// repositories
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	otpRepo := repository.NewOtpRepository(db)
	productRepo := repository.NewProductRepository(db)
	productStatusHistoryRepo := repository.NewProductStatusHistoryRepository(db)
	roleRequestRepo := repository.NewRoleRequestRepository(db)
	withdrawalRequestRepo := repository.NewWithdrawalRequestRepository(db)
	paymentMethodRepo := repository.NewPaymentMethodRepository(db)
	auctionRepo := repository.NewAuctionRepository(db)
	auctionBidRepo := repository.NewAuctionBidRepository(db)
	auctionWinnerRepo := repository.NewAuctionWinnerRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	shipmentRepo := repository.NewShipmentRepository(db)
	userAddressRepo := repository.NewUserAddressRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	userFcmTokenRepo := repository.NewUserFcmTokenRepository(db)

	repoManager := repository.NewRepositoryManager(db, userRepo, userRoleRepo, otpRepo, productRepo, productStatusHistoryRepo, roleRequestRepo, withdrawalRequestRepo, auctionRepo, auctionBidRepo, auctionWinnerRepo, paymentRepo, shipmentRepo, paymentMethodRepo, userAddressRepo, notificationRepo, userFcmTokenRepo)
	jwtInstance := internalJwt.NewJwt([]byte(config.JwtConfig.SecretKey))

	// filesystem
	fsConfig := internalFilesystem.Config{
		Filesystem: config.Filesystem.Type,
	}
	supabaseClient := infraManager.GetSupabaseStorageClient()
	if supabaseClient != nil && config.Supabase != nil {
		fsConfig.SupabaseClientConfig = &internalFilesystem.SupabaseClientConfig{
			Client:     supabaseClient,
			ProjectURL: config.Supabase.URL,
			ServiceKey: config.Supabase.ServiceKey,
			BucketName: config.Supabase.BucketName,
		}
	}
	filesystemManager := internalFilesystem.NewFilesystemManager(fsConfig)

	// use cases
	authUseCase := use_case.NewAuthUseCase(repoManager, jwtInstance)
	userUseCase := use_case.NewUserUseCase(repoManager, filesystemManager)
	userRoleUseCase := use_case.NewUserRoleUseCase(repoManager)
	productUseCase := use_case.NewProductUseCase(repoManager, filesystemManager)
	roleRequestUseCase := use_case.NewRoleRequestUseCase(repoManager, filesystemManager)
	withdrawalRequestUseCase := use_case.NewWithdrawalRequestUseCase(repoManager)
	paymentMethodUseCase := use_case.NewPaymentMethodUseCase(repoManager)
	notificationQueue := infraManager.GetNotificationQueueClient()
	auctionUseCase := use_case.NewAuctionUseCase(repoManager, filesystemManager, infraManager.GetTaskQueueClient(), notificationQueue)
	bidUseCase := use_case.NewBidUseCase(repoManager, notificationQueue, hub)
	winnerUseCase := use_case.NewWinnerUseCase(repoManager)
	paymentUseCase := use_case.NewPaymentUseCase(repoManager, infraManager.GetMidtransClient(), infraManager.GetTaskQueueClient(), infraManager.GetBiteshipClient(), notificationQueue)
	shipmentUseCase := use_case.NewShipmentUseCase(repoManager, infraManager.GetBiteshipClient(), infraManager.GetTaskQueueClient(), notificationQueue)
	userAddressUseCase := use_case.NewUserAddressUseCase(repoManager)
	biteshipUseCase := use_case.NewBiteshipUseCase(infraManager.GetBiteshipClient())
	notificationUseCase := use_case.NewNotificationUseCase(repoManager)

	baseFileUseCase := use_case.NewBaseFileUseCase(filesystemManager.Main(), filesystemManager.Tmp())

	// FIX: Urutan disesuaikan dengan parameter constructor UseCaseManager (auctionUseCase, bidUseCase, baru paymentMethodUseCase)
	ucManager := use_case.NewUseCaseManager(authUseCase, userUseCase, userRoleUseCase, productUseCase, roleRequestUseCase, withdrawalRequestUseCase, auctionUseCase, bidUseCase, paymentMethodUseCase, winnerUseCase, paymentUseCase, shipmentUseCase, userAddressUseCase, biteshipUseCase, notificationUseCase)

	return &Container{
		infrastructureManager: infraManager,
		repositoryManager:     repoManager,
		useCaseManager:        ucManager,
		filesystemManager:     filesystemManager,
		baseFileUseCase:       baseFileUseCase,
	}
}

func (c *Container) InfrastructureManager() infrastructure.InfrastructureManager {
	return c.infrastructureManager
}

func (c *Container) RepositoryManager() repository.RepositoryManager {
	return c.repositoryManager
}

func (c *Container) UseCaseManager() use_case.UseCaseManager {
	return c.useCaseManager
}

func (c *Container) FilesystemManager() internalFilesystem.FilesystemManager {
	return c.filesystemManager
}

func (c *Container) BaseFileUseCase() use_case.BaseFileUseCase {
	return c.baseFileUseCase
}

func (c *Container) Close() error {
	return c.infrastructureManager.Close()
}
