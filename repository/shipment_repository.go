package repository

import (
	"context"
	"fmt"

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
	GetById(ctx context.Context, id string) (*model.Shipment, error)
	GetByAuctionBidId(ctx context.Context, auctionBidId string) (*model.Shipment, error)
	Fetch(ctx context.Context, options ...model.ShipmentQueryOption) ([]model.Shipment, error)

	// update
	UpdateShipped(ctx context.Context, id string, trackingNumber string, courierCode string, serviceCode string, shippingCost float64, biteshipOrderId string) (*model.Shipment, error)
	UpdateReceived(ctx context.Context, id string, deliveryProofImagePath string) (*model.Shipment, error)
	UpdateBuyerAddress(ctx context.Context, id string, buyerAddressId string, snapshot string) (*model.Shipment, error)
	UpdateEstimatedCosts(ctx context.Context, id string, estimatedCosts string) (*model.Shipment, error)
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

func (r *shipmentRepository) GetById(ctx context.Context, id string) (*model.Shipment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *shipmentRepository) GetByAuctionBidId(ctx context.Context, auctionBidId string) (*model.Shipment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_bid_id"): auctionBidId}).
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

func (r *shipmentRepository) UpdateShipped(ctx context.Context, id string, trackingNumber string, courierCode string, serviceCode string, shippingCost float64, biteshipOrderId string) (*model.Shipment, error) {
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

func (r *shipmentRepository) UpdateReceived(ctx context.Context, id string, deliveryProofImagePath string) (*model.Shipment, error) {
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

func (r *shipmentRepository) UpdateBuyerAddress(ctx context.Context, id string, buyerAddressId string, snapshot string) (*model.Shipment, error) {
	now := util.CurrentDateTime()
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"buyer_address_id":       buyerAddressId,
			"buyer_address_snapshot": snapshot,
			"updated_at":             now,
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *shipmentRepository) UpdateEstimatedCosts(ctx context.Context, id string, estimatedCosts string) (*model.Shipment, error) {
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
