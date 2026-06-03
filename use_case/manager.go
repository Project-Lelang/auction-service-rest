package use_case

type UseCaseManager interface {
	AuthUseCase() AuthUseCase
	UserUseCase() UserUseCase
	UserRoleUseCase() UserRoleUseCase
	ProductUseCase() ProductUseCase
	RoleRequestUseCase() RoleRequestUseCase
	WithdrawalRequestUseCase() WithdrawalRequestUseCase
	AuctionUseCase() AuctionUseCase
	BidUseCase() BidUseCase
	WinnerUseCase() WinnerUseCase
	PaymentUseCase() PaymentUseCase
	ShipmentUseCase() ShipmentUseCase
	UserAddressUseCase() UserAddressUseCase
	BiteshipUseCase() BiteshipUseCase
}

type useCaseManager struct {
	authUseCase              AuthUseCase
	userUseCase              UserUseCase
	userRoleUseCase          UserRoleUseCase
	productUseCase           ProductUseCase
	roleRequestUseCase       RoleRequestUseCase
	withdrawalRequestUseCase WithdrawalRequestUseCase
	auctionUseCase           AuctionUseCase
	bidUseCase               BidUseCase
	winnerUseCase            WinnerUseCase
	paymentUseCase           PaymentUseCase
	shipmentUseCase          ShipmentUseCase
	userAddressUseCase       UserAddressUseCase
	biteshipUseCase          BiteshipUseCase
}

func NewUseCaseManager(
	authUseCase AuthUseCase,
	userUseCase UserUseCase,
	userRoleUseCase UserRoleUseCase,
	productUseCase ProductUseCase,
	roleRequestUseCase RoleRequestUseCase,
	withdrawalRequestUseCase WithdrawalRequestUseCase,
	auctionUseCase AuctionUseCase,
	bidUseCase BidUseCase,
	winnerUseCase WinnerUseCase,
	paymentUseCase PaymentUseCase,
	shipmentUseCase ShipmentUseCase,
	userAddressUseCase UserAddressUseCase,
	biteshipUseCase BiteshipUseCase,
) UseCaseManager {
	return &useCaseManager{
		authUseCase:              authUseCase,
		userUseCase:              userUseCase,
		userRoleUseCase:          userRoleUseCase,
		productUseCase:           productUseCase,
		roleRequestUseCase:       roleRequestUseCase,
		withdrawalRequestUseCase: withdrawalRequestUseCase,
		auctionUseCase:           auctionUseCase,
		bidUseCase:               bidUseCase,
		winnerUseCase:            winnerUseCase,
		paymentUseCase:           paymentUseCase,
		shipmentUseCase:          shipmentUseCase,
		userAddressUseCase:       userAddressUseCase,
		biteshipUseCase:          biteshipUseCase,
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

func (u *useCaseManager) BidUseCase() BidUseCase {
	return u.bidUseCase
}

func (u *useCaseManager) WinnerUseCase() WinnerUseCase {
	return u.winnerUseCase
}

func (u *useCaseManager) PaymentUseCase() PaymentUseCase {
	return u.paymentUseCase
}

func (u *useCaseManager) ShipmentUseCase() ShipmentUseCase {
	return u.shipmentUseCase
}

func (u *useCaseManager) UserAddressUseCase() UserAddressUseCase {
	return u.userAddressUseCase
}

func (u *useCaseManager) BiteshipUseCase() BiteshipUseCase {
	return u.biteshipUseCase
}
