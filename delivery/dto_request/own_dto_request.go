package dto_request

import "auction-service/data_type"

type OwnProfileUpdateRequest struct {
	Fullname string  `json:"fullname" validate:"required,max=255"            example:"John Doe"`
	Birth    string  `json:"birth"    validate:"required"                    example:"1990-01-15"`
	Gender   *string `json:"gender"   validate:"omitempty,oneof=MALE FEMALE" example:"MALE"`
} // @name OwnProfileUpdateRequest

// OwnProductFetchSorts defines sortable fields for the own product fetch endpoint.
// The anonymous-struct layout must stay identical to model.Sorts for direct
// type-conversion: model.Sorts(request.Sorts).
type OwnProductFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id name status condition created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                                  example:"desc"`
} // @name OwnProductFetchSorts

type OwnProductFetchRequest struct {
	PaginationRequest
	Sorts     OwnProductFetchSorts `json:"sorts"     validate:"dive"`
	Status    *string              `json:"status"    validate:"omitempty,oneof=DRAFT REQUEST VERIFIED REJECTED ON_BIDS COMPLETED" example:"VERIFIED"`
	Condition *string              `json:"condition" validate:"omitempty,oneof=NEW PRELOVED" example:"NEW"`
	Search    *string              `json:"search"    validate:"omitempty,max=255"                                            example:"laptop"`
} // @name OwnProductFetchRequest

type OwnProductCreateRequest struct {
	Name           string   `json:"name"             validate:"required,max=255"                    example:"Vintage Camera"`
	Description    *string  `json:"description"      validate:"omitempty,max=2000"                  example:"A beautiful vintage camera"`
	Condition      string   `json:"condition"        validate:"required,oneof=NEW PRELOVED"         example:"NEW"`
	CoverImagePath *string  `json:"cover_image_path" validate:"omitempty,max=500"                  example:"upload/abc123.jpg"`
	ImagePaths     []string `json:"image_paths"      validate:"omitempty,dive,required,max=500"`
	WeightGram     int      `json:"weight_gram"      validate:"required,gt=0"                       example:"1000"`
} // @name OwnProductCreateRequest

type OwnProductFetchStatusHistoriesRequest struct {
	ProductId int64 `json:"-" swaggerignore:"true"`
} // @name OwnProductFetchStatusHistoriesRequest

type OwnProductGetRequest struct {
	ProductId int64 `json:"-" swaggerignore:"true"`
} // @name OwnProductGetRequest

type OwnProductUpdateRequest struct {
	Name           string   `json:"name"             validate:"required,max=255"                    example:"Vintage Camera"`
	Description    *string  `json:"description"      validate:"omitempty,max=2000"                  example:"A beautiful vintage camera"`
	Condition      string   `json:"condition"        validate:"required,oneof=NEW PRELOVED"         example:"NEW"`
	CoverImagePath *string  `json:"cover_image_path" validate:"omitempty,max=500"                  example:"upload/abc123.jpg"`
	ImagePaths     []string `json:"image_paths"      validate:"omitempty,dive,required,max=500"`
	WeightGram     int      `json:"weight_gram"      validate:"required,gt=0"                       example:"1000"`

	ProductId int64 `json:"-" swaggerignore:"true"`
} // @name OwnProductUpdateRequest

type OwnProductRequestRequest struct {
	ProductId int64 `json:"-" swaggerignore:"true"`
} // @name OwnProductRequestRequest

type OwnRoleRequestCreateRequest struct {
	Role                    string  `json:"role"                      validate:"required,oneof=BIDDER SELLER" example:"BIDDER"`
	Nik                     *string `json:"nik"                       validate:"omitempty,max=50"            example:"3201010101010001"`
	IdentityImagePath       *string `json:"identity_image_path"       validate:"omitempty,max=500"           example:"/uploads/identity.jpg"`
	SelfieIdentityImagePath *string `json:"selfie_identity_image_path" validate:"omitempty,max=500"          example:"/uploads/selfie.jpg"`
	BankAccountNumber       *string `json:"bank_account_number"       validate:"omitempty,max=50"            example:"1234567890"`
} // @name OwnRoleRequestCreateRequest

type OwnWithdrawalRequestCreateRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0" example:"500000"`
} // @name OwnWithdrawalRequestCreateRequest

// OwnAuctionFetchSorts defines sortable fields for the own auction fetch endpoint.
type OwnAuctionFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id starting_price start_time end_time status fee created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                                                              example:"desc"`
} // @name OwnAuctionFetchSorts

type OwnAuctionFetchRequest struct {
	PaginationRequest
	Sorts  OwnAuctionFetchSorts `json:"sorts"  validate:"dive"`
	Status *string              `json:"status" validate:"omitempty,oneof=SCHEDULED ON_GOING WAITING_FOR_PAYMENT WAITING_FOR_SELLER_DECISION WAITING_FOR_SHIPMENT SHIPPED DELIVERED CANCELLED COMPLETED" example:"SCHEDULED"`
} // @name OwnAuctionFetchRequest

type OwnAuctionGetRequest struct {
	AuctionId int64 `json:"-" swaggerignore:"true"`
} // @name OwnAuctionGetRequest

type OwnAuctionCreateRequest struct {
	ProductId     int64              `json:"product_id"     validate:"required,gt=0" example:"1"`
	StartingPrice float64            `json:"starting_price" validate:"required,gt=0"  example:"100000"`
	StartTime     data_type.DateTime `json:"start_time"     validate:"required"        example:"2026-06-01T10:00:00Z"`
	EndTime       data_type.DateTime `json:"end_time"       validate:"required"        example:"2026-06-01T12:00:00Z"`
} // @name OwnAuctionCreateRequest

type OwnAuctionUpdateRequest struct {
	StartingPrice float64            `json:"starting_price" validate:"required,gt=0" example:"150000"`
	StartTime     data_type.DateTime `json:"start_time"     validate:"required"       example:"2026-06-01T10:00:00Z"`
	EndTime       data_type.DateTime `json:"end_time"       validate:"required"       example:"2026-06-01T12:00:00Z"`

	AuctionId int64 `json:"-" swaggerignore:"true"`
} // @name OwnAuctionUpdateRequest

// OwnBidFetchSorts defines sortable fields for the own bid fetch endpoint.
type OwnBidFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id amount created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                       example:"desc"`
} // @name OwnBidFetchSorts

type OwnBidFetchRequest struct {
	PaginationRequest
	Sorts     OwnBidFetchSorts `json:"sorts"      validate:"dive"`
	AuctionId *int64           `json:"auction_id" validate:"omitempty,gt=0" example:"1"`
} // @name OwnBidFetchRequest

type OwnBidGetRequest struct {
	BidId int64 `json:"-" swaggerignore:"true"`
} // @name OwnBidGetRequest

// OwnPaymentFetchSorts defines sortable fields for the own payment fetch endpoint.
type OwnPaymentFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id amount status created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                              example:"desc"`
} // @name OwnPaymentFetchSorts

type OwnPaymentFetchRequest struct {
	PaginationRequest
	Sorts  OwnPaymentFetchSorts `json:"sorts"  validate:"dive"`
	Status *string              `json:"status" validate:"omitempty,oneof=WAITING_FOR_PAYMENT PAID FAILED EXPIRED CANCELLED" example:"WAITING_FOR_PAYMENT"`
} // @name OwnPaymentFetchRequest

type OwnPaymentGetRequest struct {
	PaymentId int64 `json:"-" swaggerignore:"true"`
} // @name OwnPaymentGetRequest

type OwnNotificationFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id title type is_read created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                                  example:"desc"`
} // @name OwnNotificationFetchSorts

type OwnNotificationFetchRequest struct {
	PaginationRequest
	Sorts  OwnNotificationFetchSorts `json:"sorts"   validate:"dive"`
	IsRead *bool                     `json:"is_read" validate:"omitempty" example:"false"`
} // @name OwnNotificationFetchRequest

type OwnNotificationGetRequest struct {
	NotificationId int64 `json:"-" swaggerignore:"true"`
} // @name OwnNotificationGetRequest

type OwnNotificationMarkReadRequest struct {
	NotificationId int64 `json:"-" swaggerignore:"true"`
} // @name OwnNotificationMarkReadRequest
