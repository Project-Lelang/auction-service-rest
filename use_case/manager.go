package use_case

type UseCaseManager interface {
	AuthUseCase() AuthUseCase
	UserUseCase() UserUseCase
	UserRoleUseCase() UserRoleUseCase
	ProductUseCase() ProductUseCase
	RoleRequestUseCase() RoleRequestUseCase
	WithdrawalRequestUseCase() WithdrawalRequestUseCase
	AuctionUseCase() AuctionUseCase
}

type useCaseManager struct {
	authUseCase              AuthUseCase
	userUseCase              UserUseCase
	userRoleUseCase          UserRoleUseCase
	productUseCase           ProductUseCase
	roleRequestUseCase       RoleRequestUseCase
	withdrawalRequestUseCase WithdrawalRequestUseCase
	auctionUseCase           AuctionUseCase
}

func NewUseCaseManager(
	authUseCase AuthUseCase,
	userUseCase UserUseCase,
	userRoleUseCase UserRoleUseCase,
	productUseCase ProductUseCase,
	roleRequestUseCase RoleRequestUseCase,
	withdrawalRequestUseCase WithdrawalRequestUseCase,
	auctionUseCase AuctionUseCase,
) UseCaseManager {
	return &useCaseManager{
		authUseCase:              authUseCase,
		userUseCase:              userUseCase,
		userRoleUseCase:          userRoleUseCase,
		productUseCase:           productUseCase,
		roleRequestUseCase:       roleRequestUseCase,
		withdrawalRequestUseCase: withdrawalRequestUseCase,
		auctionUseCase:           auctionUseCase,
	}
}

func (u *useCaseManager) AuthUseCase() AuthUseCase {
	return u.authUseCase
}

func (u *useCaseManager) UserUseCase() UserUseCase {
	return u.userUseCase
}

func (u *useCaseManager) UserRoleUseCase() UserRoleUseCase {
	return u.userRoleUseCase
}

func (u *useCaseManager) ProductUseCase() ProductUseCase {
	return u.productUseCase
}

func (u *useCaseManager) RoleRequestUseCase() RoleRequestUseCase {
	return u.roleRequestUseCase
}

func (u *useCaseManager) WithdrawalRequestUseCase() WithdrawalRequestUseCase {
	return u.withdrawalRequestUseCase
}

func (u *useCaseManager) AuctionUseCase() AuctionUseCase {
	return u.auctionUseCase
}
