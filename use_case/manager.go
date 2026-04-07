package use_case

type UseCaseManager interface {
	AuthUseCase() AuthUseCase
}

type useCaseManager struct {
	authUseCase AuthUseCase
}

func NewUseCaseManager(
	authUseCase AuthUseCase,
) UseCaseManager {
	return &useCaseManager{
		authUseCase: authUseCase,
	}
}

func (u *useCaseManager) AuthUseCase() AuthUseCase {
	return u.authUseCase
}
