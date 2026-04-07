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
	loggerStack := infraManager.GetLoggerStack()

	// repositories
	userRepo := repository.NewUserRepository(db, loggerStack)
	userAccessTokenRepo := repository.NewUserAccessTokenRepository(db, loggerStack)

	repoManager := repository.NewRepositoryManager(userRepo, userAccessTokenRepo)

	// jwt
	jwtInstance := internalJwt.NewJwt([]byte(config.JwtConfig.SecretKey))

	// use cases
	authUseCase := use_case.NewAuthUseCase(repoManager, jwtInstance)
	ucManager := use_case.NewUseCaseManager(authUseCase)

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
