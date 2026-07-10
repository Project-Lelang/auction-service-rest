package repository

import (
	"context"
	"fmt"
	"strings"

	"auction-service/constant"
	"auction-service/data_type"
	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

// ShipmentRepository defines persistence operations for shipments.
type ShipmentRepository interface {
	// create
	Insert(ctx context.Context, shipment *model.Shipment) error

	// read
	GetById(ctx context.Context, id int64) (*model.Shipment, error)
	GetByIdForUpdate(ctx context.Context, id int64) (*model.Shipment, error)
	GetByAuctionBidId(ctx context.Context, auctionBidId int64) (*model.Shipment, error)
	GetByTrackingIdentifier(ctx context.Context, identifiers []string) (*model.Shipment, error)
	Fetch(ctx context.Context, options ...model.ShipmentQueryOption) ([]model.Shipment, error)
	FetchPendingAddressDeadline(ctx context.Context) ([]model.Shipment, error)
	FetchPendingShipDeadline(ctx context.Context) ([]model.Shipment, error)
	FetchPendingDeliveryTracking(ctx context.Context) ([]model.Shipment, error)
	FetchPendingReceiveDeadline(ctx context.Context) ([]model.Shipment, error)

	// update
	UpdateShipped(ctx context.Context, id int64, trackingNumber string, courierCode string, serviceCode string, shippingCost float64, biteshipOrderId string) (*model.Shipment, error)
	UpdateReceived(ctx context.Context, id int64, deliveryProofImagePath string) (*model.Shipment, error)
	UpdateAutoReceived(ctx context.Context, id int64) (*model.Shipment, error)
	UpdateBidderAddressFailed(ctx context.Context, id int64) (*model.Shipment, error)
	UpdateSellerFailed(ctx context.Context, id int64) (*model.Shipment, error)
	UpdateBidderAddressDeadline(ctx context.Context, id int64, bidderAddressDeadlineAt data_type.DateTime) (*model.Shipment, error)
	UpdateShipDeadline(ctx context.Context, id int64, shipDeadlineAt data_type.DateTime) (*model.Shipment, error)
	UpdateDelivered(ctx context.Context, id int64, receiveDeadlineAt data_type.DateTime) (*model.Shipment, error)
	UpdateBidderAddress(ctx context.Context, id int64, bidderAddressId int64, snapshot string) (*model.Shipment, error)
	UpdateSellerAddress(ctx context.Context, id int64, sellerAddressId int64, snapshot string) (*model.Shipment, error)
	UpdateEstimatedCosts(ctx context.Context, id int64, estimatedCosts string) (*model.Shipment, error)
}

type shipmentRepository struct {
	db infrastructure.DBTX
}

func NewShipmentRepository(db infrastructure.DBTX) ShipmentRepository {
	return &shipmentRepository{db: db}
}

func (r *shipmentRepository) tableName() string { return model.ShipmentTableName }
func (r *shipmentRepository) alias() string     { return "s" }
func (r *shipmentRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *shipmentRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *shipmentRepository) buildBaseStmt(option model.ShipmentQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().From(r.fromTable())

	if option.AuctionBidId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("auction_bid_id"): *option.AuctionBidId})
	}
	if option.UserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("user_id"): *option.UserId})
	}

	return stmt
}

func (r *shipmentRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.Shipment, error) {
	s := model.Shipment{}
	if err := get(r.db, ctx, &s, stmt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shipmentRepository) fetchInternal(ctx context.Context, stmt squirrel.SelectBuilder) ([]model.Shipment, error) {
	shipments := []model.Shipment{}
	if err := fetch(r.db, ctx, &shipments, stmt); err != nil {
		return nil, err
	}
	return shipments, nil
}

func (r *shipmentRepository) Insert(ctx context.Context, shipment *model.Shipment) error {
	return defaultInsert(r.db, ctx, shipment)
}

func (r *shipmentRepository) GetById(ctx context.Context, id int64) (*model.Shipment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *shipmentRepository) GetByIdForUpdate(ctx context.Context, id int64) (*model.Shipment, error) {
	query := fmt.Sprintf(
		"SELECT %s.* FROM %s WHERE %s.id = ? LIMIT 1 FOR UPDATE",
		r.alias(), r.fromTable(), r.alias(),
	)
	s := model.Shipment{}
	dt := dbtx(r.db, ctx)
	if err := dt.GetContext(ctx, &s, query, id); err != nil {
		return nil, translateSqlError(err)
	}
	return &s, nil
}

func (r *shipmentRepository) GetByAuctionBidId(ctx context.Context, auctionBidId int64) (*model.Shipment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_bid_id"): auctionBidId}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *shipmentRepository) GetByTrackingIdentifier(ctx context.Context, identifiers []string) (*model.Shipment, error) {
	clean := make([]string, 0, len(identifiers))
	seen := map[string]struct{}{}
	for _, identifier := range identifiers {
		identifier = strings.TrimSpace(identifier)
		if identifier == "" {
			continue
		}
		if _, ok := seen[identifier]; ok {
			continue
		}
		seen[identifier] = struct{}{}
		clean = append(clean, identifier)
	}
	if len(clean) == 0 {
		return nil, constant.ErrNoData
	}
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Or{
			squirrel.Eq{r.f("tracking_number"): clean},
			squirrel.Eq{r.f("biteship_order_id"): clean},
		}).
		OrderBy(r.f("created_at") + " DESC").
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *shipmentRepository) Fetch(ctx context.Context, options ...model.ShipmentQueryOption) ([]model.Shipment, error) {
	option := model.ShipmentQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}
	stmt := r.buildBaseStmt(option).Column(r.f("*"))
	stmt = model.Prepare(stmt, &option)
	return r.fetchInternal(ctx, stmt)
}

func (r *shipmentRepository) FetchPendingAddressDeadline(ctx context.Context) ([]model.Shipment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Join("auction_bids ab ON ab.id = " + r.f("auction_bid_id")).
		Join("auctions a ON a.id = ab.auction_id").
		Where(squirrel.Eq{"a.status": constant.AuctionStatusWaitingForBidderAddress}).
		Where(squirrel.Expr(r.f("bidder_address_failed_at") + " IS NULL")).
		Where(squirrel.Expr(r.f("bidder_address_deadline_at") + " IS NOT NULL")).
		OrderBy(r.f("bidder_address_deadline_at") + " ASC")
	return r.fetchInternal(ctx, stmt)
}

func (r *shipmentRepository) FetchPendingShipDeadline(ctx context.Context) ([]model.Shipment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Join("auction_bids ab ON ab.id = " + r.f("auction_bid_id")).
		Join("auctions a ON a.id = ab.auction_id").
		Where(squirrel.Eq{"a.status": constant.AuctionStatusWaitingForShipment}).
		Where(squirrel.Expr(r.f("shipped_at") + " IS NULL")).
		Where(squirrel.Expr(r.f("seller_failed_at") + " IS NULL")).
		Where(squirrel.Expr(r.f("ship_deadline_at") + " IS NOT NULL")).
		OrderBy(r.f("ship_deadline_at") + " ASC")
	return r.fetchInternal(ctx, stmt)
}

func (r *shipmentRepository) FetchPendingReceiveDeadline(ctx context.Context) ([]model.Shipment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Join("auction_bids ab ON ab.id = " + r.f("auction_bid_id")).
		Join("auctions a ON a.id = ab.auction_id").
		Where(squirrel.Eq{"a.status": constant.AuctionStatusShipped}).
		Where(squirrel.Expr(r.f("shipped_at") + " IS NOT NULL")).
		Where(squirrel.Expr(r.f("received_at") + " IS NULL")).
		Where(squirrel.Expr(r.f("seller_failed_at") + " IS NULL")).
		Where(squirrel.Expr(r.f("receive_deadline_at") + " IS NOT NULL")).
		OrderBy(r.f("receive_deadline_at") + " ASC")
	return r.fetchInternal(ctx, stmt)
}

func (r *shipmentRepository) FetchPendingDeliveryTracking(ctx context.Context) ([]model.Shipment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Join("auction_bids ab ON ab.id = " + r.f("auction_bid_id")).
		Join("auctions a ON a.id = ab.auction_id").
		Where(squirrel.Eq{"a.status": constant.AuctionStatusShipped}).
		Where(squirrel.Expr(r.f("tracking_number") + " IS NOT NULL")).
		Where(squirrel.Expr(r.f("shipped_at") + " IS NOT NULL")).
		Where(squirrel.Expr(r.f("delivered_at") + " IS NULL")).
		Where(squirrel.Expr(r.f("receive_deadline_at") + " IS NULL")).
		Where(squirrel.Expr(r.f("received_at") + " IS NULL")).
		Where(squirrel.Expr(r.f("seller_failed_at") + " IS NULL")).
		OrderBy(r.f("shipped_at") + " ASC")
	return r.fetchInternal(ctx, stmt)
}

func (r *shipmentRepository) UpdateShipped(ctx context.Context, id int64, trackingNumber string, courierCode string, serviceCode string, shippingCost float64, biteshipOrderId string) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"tracking_number":   trackingNumber,
			"courier_code":      courierCode,
			"service_code":      serviceCode,
			"shipping_cost":     shippingCost,
			"biteship_order_id": biteshipOrderId,
			"shipped_at":        now,
			"updated_at":        now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *shipmentRepository) UpdateDelivered(ctx context.Context, id int64, receiveDeadlineAt data_type.DateTime) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"delivered_at":        now,
			"receive_deadline_at": receiveDeadlineAt,
			"updated_at":          now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *shipmentRepository) UpdateAutoReceived(ctx context.Context, id int64) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"received_at":      now,
			"auto_received_at": now,
			"updated_at":       now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *shipmentRepository) UpdateBidderAddressFailed(ctx context.Context, id int64) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"bidder_address_failed_at": now,
			"updated_at":               now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *shipmentRepository) UpdateSellerFailed(ctx context.Context, id int64) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"seller_failed_at": now,
			"updated_at":       now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *shipmentRepository) UpdateBidderAddressDeadline(ctx context.Context, id int64, bidderAddressDeadlineAt data_type.DateTime) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"bidder_address_deadline_at": bidderAddressDeadlineAt,
			"updated_at":                 now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *shipmentRepository) UpdateShipDeadline(ctx context.Context, id int64, shipDeadlineAt data_type.DateTime) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"ship_deadline_at": shipDeadlineAt,
			"updated_at":       now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *shipmentRepository) UpdateReceived(ctx context.Context, id int64, deliveryProofImagePath string) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"delivery_proof_image_path": deliveryProofImagePath,
			"received_at":               now,
			"updated_at":                now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *shipmentRepository) UpdateBidderAddress(ctx context.Context, id int64, bidderAddressId int64, snapshot string) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"bidder_address_id":       bidderAddressId,
			"bidder_address_snapshot": snapshot,
			"updated_at":              now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *shipmentRepository) UpdateSellerAddress(ctx context.Context, id int64, sellerAddressId int64, snapshot string) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"seller_address_id":       sellerAddressId,
			"seller_address_snapshot": snapshot,
			"updated_at":              now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *shipmentRepository) UpdateEstimatedCosts(ctx context.Context, id int64, estimatedCosts string) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"estimated_costs": estimatedCosts,
			"updated_at":      now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}
