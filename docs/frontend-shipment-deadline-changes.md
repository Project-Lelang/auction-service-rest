# Frontend Shipment Deadline Changes

## Summary

Backend now resolves two shipment corner cases automatically:

- Buyer does not confirm/select shipping address after payment.
- Seller does not ship after buyer confirms address.
- Buyer does not confirm receipt after Biteship webhook/tracking confirms delivered.

The deadlines are stored on the shipment record and returned by the existing shipment APIs.

## Existing Endpoints Affected

- `POST /auctions/{auctionId}/shipments/filter`
- `GET /auctions/{auctionId}/shipments/{shipmentId}`
- `PATCH /auctions/{auctionId}/shipments/{shipmentId}/buyer-address`
- `POST /auctions/{auctionId}/shipments/{shipmentId}/ship`
- `POST /auctions/{auctionId}/shipments/{shipmentId}/receive`

No new frontend endpoint is required.

## New Shipment Response Fields

`ShipmentResponse` now includes:

```json
{
  "buyer_address_deadline_at": "2026-06-27T10:00:00+07:00",
  "ship_deadline_at": "2026-06-28T10:00:00+07:00",
  "delivered_at": "2026-07-01T10:00:00+07:00",
  "receive_deadline_at": "2026-07-05T10:00:00+07:00",
  "auto_received_at": null,
  "buyer_address_failed_at": null,
  "seller_failed_at": null
}
```

Field meaning:

- `buyer_address_deadline_at`: buyer must confirm/select shipping address before this time. Set after payment completes.
- `ship_deadline_at`: seller must ship before this time. Set after buyer confirms address.
- `delivered_at`: filled when Biteship webhook or tracking fallback confirms the package is delivered/received.
- `receive_deadline_at`: buyer must confirm receipt before this time. Set only after `delivered_at` is filled.
- `auto_received_at`: filled when backend auto-completes because buyer did not confirm in time.
- `buyer_address_failed_at`: filled when backend refunds because buyer did not confirm/select address in time.
- `seller_failed_at`: filled when backend marks seller failed to ship.

## Status Changes To Handle

Auction status can now become:

```text
SELLER_FAILED_TO_SHIP
```

Payment status can now become:

```text
REFUNDED
```

Product status can now become:

```text
SELLER_FAILED_TO_SHIP
```

Existing statuses still apply:

```text
WAITING_FOR_BUYER_ADDRESS
WAITING_FOR_SELLER_DECISION
WAITING_FOR_SHIPMENT
SHIPPED
COMPLETED
```

## Suggested UI Behavior

When auction status is `WAITING_FOR_BUYER_ADDRESS`:

- Show buyer address confirmation deadline from `shipment.buyer_address_deadline_at`.
- Buyer can still select/confirm address before the deadline.
- Seller should see that the buyer is selecting address.

When auction status is `WAITING_FOR_SELLER_DECISION` and `shipment.buyer_address_failed_at` is not null:

- Disable address confirmation, ship, and receive actions.
- Buyer should see refund information.
- Seller should see relist/second-chance actions if the frontend already supports them.

When auction status is `WAITING_FOR_SHIPMENT`:

- Show seller shipment deadline from `shipment.ship_deadline_at`.
- Seller can still ship before the deadline.
- Buyer should see that the seller is preparing shipment.

When auction status is `SELLER_FAILED_TO_SHIP`:

- Disable ship and receive actions.
- Buyer should see refund information.
- Seller should see that shipment deadline was missed.

When auction status is `SHIPPED`:

- If `shipment.receive_deadline_at` is null, show shipment is in transit and waiting for courier delivery confirmation.
- If `shipment.receive_deadline_at` is not null, show buyer confirmation deadline.
- Buyer can click receive after shipment is sent, but backend auto-complete only starts after Biteship tracking confirms delivery.
- Seller should see that payment will release automatically after delivery confirmation deadline passes.

When auction status is `COMPLETED` and `shipment.auto_received_at` is not null:

- Show the order as auto-completed by system.
- Do not require delivery proof image.

## Notification Events

Backend can now send these FCM/database notification types:

```text
BUYER_ADDRESS_EXPIRED
SHIPMENT_REFUNDED
SHIPMENT_AUTO_COMPLETED
SHIPMENT_DEADLINE_MISSED
SHIPMENT_DELIVERED
```

## Biteship Webhook

Backend exposes a public webhook endpoint:

```text
POST /biteship/webhooks
```

When Biteship sends a delivered/received/completed tracking event, backend resolves the shipment by tracking/waybill fields, sets `delivered_at`, sets `receive_deadline_at`, and sends `SHIPMENT_DELIVERED`.

Polling remains as a fallback if a webhook is delayed or missed.

Use the same auction-detail redirect behavior as existing auction notifications. Payload includes:

```json
{
  "auction_id": "123",
  "auction_url": "http://localhost:3000/auction/123"
}
```

## Current Deadline Defaults

Configured in `conf.yml`:

```yaml
shipment_deadline:
  buyer_address_hours: 24
  seller_ship_hours: 72
  buyer_receive_hours: 168
  tracking_check_interval_minutes: 60
  deadline_grace_minutes: 5
```

This means:

- Buyer has 24 hours after payment completion to confirm/select address.
- Seller has 72 hours after buyer confirms address.
- Backend prefers Biteship webhook, and also checks Biteship tracking every 60 minutes after seller ships until delivered.
- Buyer has 168 hours after Biteship tracking confirms delivered.
- Redis scheduled tasks run with a 5 minute grace window.
