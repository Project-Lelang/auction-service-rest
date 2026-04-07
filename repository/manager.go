package repository

type RepositoryManager interface {
	UserRepository() UserRepository
	UserAccessTokenRepository() UserAccessTokenRepository
}

type repositoryManager struct {
	userRepository            UserRepository
	userAccessTokenRepository UserAccessTokenRepository
}

func NewRepositoryManager(
	userRepository UserRepository,
	userAccessTokenRepository UserAccessTokenRepository,
) RepositoryManager {
	return &repositoryManager{
		userRepository:            userRepository,
		userAccessTokenRepository: userAccessTokenRepository,
	}
}

func (r *repositoryManager) UserRepository() UserRepository {
	return r.userRepository
}

func (r *repositoryManager) UserAccessTokenRepository() UserAccessTokenRepository {
	return r.userAccessTokenRepository
}
