package manager

import (
	"encoding/json"
	"os"

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

func NewContainer() *Container {
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
	auctionRepo := repository.NewAuctionRepository(db)

	repoManager := repository.NewRepositoryManager(db, userRepo, userRoleRepo, otpRepo, productRepo, productStatusHistoryRepo, roleRequestRepo, withdrawalRequestRepo, auctionRepo)
	jwtInstance := internalJwt.NewJwt([]byte(config.JwtConfig.SecretKey))

	// filesystem
	fsConfig := internalFilesystem.Config{
		Filesystem: config.Filesystem.Type,
	}
	gcsClient := infraManager.GetGcsClient()
	if gcsClient != nil && config.Gcs != nil {
		gcsClientConfig := &internalFilesystem.GcsClientConfig{
			Client:     gcsClient,
			ProjectId:  config.Gcs.ProjectId,
			BucketName: config.Gcs.BucketName,
		}
		// Load service account credentials for signed URL generation
		if jsonBytes, err := os.ReadFile(config.Gcs.ConfigFilepath()); err == nil {
			var sa struct {
				ClientEmail string `json:"client_email"`
				PrivateKey  string `json:"private_key"`
			}
			if err := json.Unmarshal(jsonBytes, &sa); err == nil {
				gcsClientConfig.ClientEmail = sa.ClientEmail
				gcsClientConfig.PrivateKey = sa.PrivateKey
			}
		}
		fsConfig.GcsClientConfig = gcsClientConfig
	}
	filesystemManager := internalFilesystem.NewFilesystemManager(fsConfig)

	// use cases
	authUseCase := use_case.NewAuthUseCase(repoManager, jwtInstance)
	userUseCase := use_case.NewUserUseCase(repoManager, filesystemManager)
	userRoleUseCase := use_case.NewUserRoleUseCase(repoManager)
	productUseCase := use_case.NewProductUseCase(repoManager, filesystemManager)
	roleRequestUseCase := use_case.NewRoleRequestUseCase(repoManager, filesystemManager)
	withdrawalRequestUseCase := use_case.NewWithdrawalRequestUseCase(repoManager)
	auctionUseCase := use_case.NewAuctionUseCase(repoManager)

	baseFileUseCase := use_case.NewBaseFileUseCase(filesystemManager.Main(), filesystemManager.Tmp())

	ucManager := use_case.NewUseCaseManager(authUseCase, userUseCase, userRoleUseCase, productUseCase, roleRequestUseCase, withdrawalRequestUseCase, auctionUseCase)

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
	return c.infrastructureManager.CloseDB()
}
