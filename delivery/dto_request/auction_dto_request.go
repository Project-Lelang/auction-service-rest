package dto_request

// --------------------------------------------------------------------- auction

type AuctionFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id starting_price start_time end_time status fee created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                                                              example:"desc"`
} // @name AuctionFetchSorts

type AuctionFetchRequest struct {
	PaginationRequest
	Sorts  AuctionFetchSorts `json:"sorts"  validate:"dive"`
	Status *string           `json:"status" validate:"omitempty,oneof=SCHEDULED ON_GOING WAITING_FOR_PAYMENT WAITING_FOR_SELLER_DECISION WAITING_FOR_SHIPMENT SHIPPED DELIVERED CANCELLED COMPLETED" example:"ON_GOING"`
} // @name AuctionFetchRequest

type AuctionGetRequest struct {
	AuctionId int64 `json:"-" swaggerignore:"true"`
} // @name AuctionGetRequest

type OwnAuctionRelistRequest struct {
	AuctionId int64 `json:"-" swaggerignore:"true"`
} // @name OwnAuctionRelistRequest

type OwnAuctionSecondChanceRequest struct {
	AuctionId int64 `json:"-" swaggerignore:"true"`
} // @name OwnAuctionSecondChanceRequest

// --------------------------------------------------------------------- bid

type AuctionBidCreateRequest struct {
	Amount    float64 `json:"amount"     validate:"required,gt=0" example:"150000"`
	AuctionId int64   `json:"-"           swaggerignore:"true"`
} // @name AuctionBidCreateRequest

type AuctionBidFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id amount created_at updated_at" example:"amount"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                       example:"desc"`
} // @name AuctionBidFetchSorts

type AuctionBidFetchRequest struct {
	PaginationRequest
	Sorts     AuctionBidFetchSorts `json:"sorts" validate:"dive"`
	AuctionId int64                `json:"-"     swaggerignore:"true"`
} // @name AuctionBidFetchRequest

// --------------------------------------------------------------------- winner

type AuctionWinnerFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id status created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                       example:"desc"`
} // @name AuctionWinnerFetchSorts

type AuctionWinnerFetchRequest struct {
	PaginationRequest
	Sorts     AuctionWinnerFetchSorts `json:"sorts" validate:"dive"`
	AuctionId int64                   `json:"-"     swaggerignore:"true"`
} // @name AuctionWinnerFetchRequest

type AuctionWinnerGetRequest struct {
	AuctionId int64 `json:"-" swaggerignore:"true"`
	WinnerId  int64 `json:"-" swaggerignore:"true"`
} // @name AuctionWinnerGetRequest

// --------------------------------------------------------------------- payment

type AuctionPaymentGetRequest struct {
	AuctionId int64 `json:"-" swaggerignore:"true"`
	PaymentId int64 `json:"-" swaggerignore:"true"`
} // @name AuctionPaymentGetRequest

// --------------------------------------------------------------------- shipment

type AuctionShipmentFetchRequest struct {
	AuctionId int64 `json:"-" swaggerignore:"true"`
} // @name AuctionShipmentFetchRequest

type AuctionShipmentGetRequest struct {
	AuctionId  int64 `json:"-" swaggerignore:"true"`
	ShipmentId int64 `json:"-" swaggerignore:"true"`
} // @name AuctionShipmentGetRequest

type AuctionShipmentShipRequest struct {
	CourierCode string `json:"courier_code" validate:"required,max=50" example:"jne"`
	ServiceCode string `json:"service_code" validate:"required,max=20" example:"reg"`
	AuctionId   int64  `json:"-"            swaggerignore:"true"`
	ShipmentId  int64  `json:"-"            swaggerignore:"true"`
} // @name AuctionShipmentShipRequest

type AuctionShipmentReceiveRequest struct {
	DeliveryProofImagePath string `json:"delivery_proof_image_path" validate:"required,max=500" example:"/uploads/proof.jpg"`
	AuctionId              int64  `json:"-"                         swaggerignore:"true"`
	ShipmentId             int64  `json:"-"                         swaggerignore:"true"`
} // @name AuctionShipmentReceiveRequest

type AuctionShipmentUpdateAddressRequest struct {
	AddressId  int64 `json:"address_id" validate:"required,gt=0" example:"1"`
	AuctionId  int64 `json:"-"          swaggerignore:"true"`
	ShipmentId int64 `json:"-"          swaggerignore:"true"`
} // @name AuctionShipmentUpdateAddressRequest

type AuctionShipmentGetTrackingRequest struct {
	AuctionId  int64 `json:"-" swaggerignore:"true"`
	ShipmentId int64 `json:"-" swaggerignore:"true"`
} // @name AuctionShipmentGetTrackingRequest
