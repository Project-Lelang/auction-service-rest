package repository

type RepositoryManager interface {
	UserRepository() UserRepository
	UserRoleRepository() UserRoleRepository
	OtpRepository() OtpRepository
	ProductRepository() ProductRepository
	ProductStatusHistoryRepository() ProductStatusHistoryRepository
}

type repositoryManager struct {
	userRepository                 UserRepository
	userRoleRepository             UserRoleRepository
	otpRepository                  OtpRepository
	productRepository              ProductRepository
	productStatusHistoryRepository ProductStatusHistoryRepository
}

func NewRepositoryManager(
	userRepository UserRepository,
	userRoleRepository UserRoleRepository,
	otpRepository OtpRepository,
	productRepository ProductRepository,
	productStatusHistoryRepository ProductStatusHistoryRepository,
) RepositoryManager {
	return &repositoryManager{
		userRepository:                 userRepository,
		userRoleRepository:             userRoleRepository,
		otpRepository:                  otpRepository,
		productRepository:              productRepository,
		productStatusHistoryRepository: productStatusHistoryRepository,
	}
}

func (r *repositoryManager) UserRepository() UserRepository {
	return r.userRepository
}

func (r *repositoryManager) UserRoleRepository() UserRoleRepository {
	return r.userRoleRepository
}

func (r *repositoryManager) OtpRepository() OtpRepository {
	return r.otpRepository
}

func (r *repositoryManager) ProductRepository() ProductRepository {
	return r.productRepository
}

func (r *repositoryManager) ProductStatusHistoryRepository() ProductStatusHistoryRepository {
	return r.productStatusHistoryRepository
}
