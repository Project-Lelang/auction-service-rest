package manager

import (
	"auction-service/global"
	"auction-service/infrastructure"
	internalJwt "auction-service/internal/jwt"
	"auction-service/repository"
	"auction-service/use_case"
)

type Container struct {
	infrastructureManager infrastructure.InfrastructureManager
	repositoryManager     repository.RepositoryManager
	useCaseManager        use_case.UseCaseManager
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

	repoManager := repository.NewRepositoryManager(userRepo, userRoleRepo, otpRepo, productRepo, productStatusHistoryRepo)

	// jwt
	jwtInstance := internalJwt.NewJwt([]byte(config.JwtConfig.SecretKey))

	// use cases
	authUseCase := use_case.NewAuthUseCase(repoManager, jwtInstance)
	userUseCase := use_case.NewUserUseCase(repoManager)
	userRoleUseCase := use_case.NewUserRoleUseCase(repoManager)
	productUseCase := use_case.NewProductUseCase(repoManager)

	ucManager := use_case.NewUseCaseManager(authUseCase, userUseCase, userRoleUseCase, productUseCase)

	return &Container{
		infrastructureManager: infraManager,
		repositoryManager:     repoManager,
		useCaseManager:        ucManager,
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

func (c *Container) Close() error {
	return c.infrastructureManager.CloseDB()
}
