package infrastructure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const biteshipBaseURL = "https://api.biteship.com"

// --------------------------------------------------------------------- domain types

// BiteshipArea represents a single area/subdistrict returned by the Maps API.
type BiteshipArea struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
	Province    string `json:"administrative_division_level_1_name"`
	City        string `json:"administrative_division_level_2_name"`
	District    string `json:"administrative_division_level_3_name"`
	PostalCode  int    `json:"postal_code"`
}

// BiteshipShippingOption represents a single courier/service pricing option.
type BiteshipShippingOption struct {
	CourierName        string `json:"courier_name"`
	CourierCode        string `json:"courier_code"`
	CourierServiceName string `json:"courier_service_name"`
	CourierServiceCode string `json:"courier_service_code"`
	ShippingFee        int    `json:"shipping_fee"`
	Price              int    `json:"price"`
	Duration           string `json:"duration"`
}

// BiteshipOrderResult is the outcome of CreateOrder.
type BiteshipOrderResult struct {
	OrderId    string // Biteship internal order ID
	WaybillId  string // AWB number
	TrackingId string // Biteship tracking ID (used for GetTracking)
}

// BiteshipOrderItem describes one product in an order.
type BiteshipOrderItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Value       int    `json:"value"`
	Weight      int    `json:"weight"` // grams
	Quantity    int    `json:"quantity"`
	Length      int    `json:"length"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

// BiteshipCreateOrderRequest is the request body for POST /v1/orders.
type BiteshipCreateOrderRequest struct {
	OriginContactName       string              `json:"origin_contact_name"`
	OriginContactPhone      string              `json:"origin_contact_phone"`
	OriginAddress           string              `json:"origin_address"`
	OriginAreaId            string              `json:"origin_area_id"`
	DestinationContactName  string              `json:"destination_contact_name"`
	DestinationContactPhone string              `json:"destination_contact_phone"`
	DestinationAddress      string              `json:"destination_address"`
	DestinationAreaId       string              `json:"destination_area_id"`
	CourierCompany          string              `json:"courier_company"`
	CourierType             string              `json:"courier_type"`
	DeliveryType            string              `json:"delivery_type"` // "now" or "scheduled"
	Items                   []BiteshipOrderItem `json:"items"`
}

// --------------------------------------------------------------------- client interface

// BiteshipClient abstracts calls to the Biteship shipping API.
type BiteshipClient interface {
	// SearchAreas searches Biteship area IDs by keyword (city/district/postal code).
	SearchAreas(keyword string) ([]BiteshipArea, error)

	// Calculate returns shipping rate options for origin→destination with given items.
	Calculate(originAreaId, destAreaId string, weightGrams, itemValue int) ([]BiteshipShippingOption, error)

	// CreateOrder creates a delivery order; returns order details including AWB and tracking ID.
	CreateOrder(req BiteshipCreateOrderRequest) (BiteshipOrderResult, error)

	// GetTracking returns the full tracking detail for a Biteship tracking ID.
	GetTracking(trackingId string) (map[string]interface{}, error)
}

// --------------------------------------------------------------------- implementation

type biteshipClient struct {
	apiKey string
	http   *http.Client
}

func NewBiteshipClient(apiKey string) BiteshipClient {
	return &biteshipClient{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *biteshipClient) do(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, biteshipBaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *biteshipClient) SearchAreas(keyword string) ([]BiteshipArea, error) {
	path := "/v1/maps/areas?countries=ID&input=" + url.QueryEscape(keyword) + "&type=single"
	raw, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Success bool           `json:"success"`
		Error   string         `json:"error"`
		Areas   []BiteshipArea `json:"areas"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("biteship SearchAreas failed: %s", resp.Error)
	}
	return resp.Areas, nil
}

func (c *biteshipClient) Calculate(originAreaId, destAreaId string, weightGrams, itemValue int) ([]BiteshipShippingOption, error) {
	body := map[string]interface{}{
		"origin_area_id":      originAreaId,
		"destination_area_id": destAreaId,
		"couriers":            "jne,jnt,sicepat,anteraja,wahana,idexpress",
		"items": []map[string]interface{}{
			{
				"name":     "item",
				"value":    itemValue,
				"weight":   weightGrams,
				"quantity": 1,
				"length":   10,
				"width":    10,
				"height":   10,
			},
		},
	}
	raw, err := c.do("POST", "/v1/rates/couriers", body)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
		Pricing []struct {
			CourierName        string `json:"courier_name"`
			CourierCode        string `json:"courier_code"`
			CourierServiceName string `json:"courier_service_name"`
			CourierServiceCode string `json:"courier_service_code"`
			ShippingFee        int    `json:"shipping_fee"`
			Price              int    `json:"price"`
			Duration           string `json:"duration"`
		} `json:"pricing"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("biteship Calculate failed: %s", resp.Error)
	}
	options := make([]BiteshipShippingOption, len(resp.Pricing))
	for i, p := range resp.Pricing {
		options[i] = BiteshipShippingOption{
			CourierName:        p.CourierName,
			CourierCode:        p.CourierCode,
			CourierServiceName: p.CourierServiceName,
			CourierServiceCode: p.CourierServiceCode,
			ShippingFee:        p.ShippingFee,
			Price:              p.Price,
			Duration:           p.Duration,
		}
	}
	return options, nil
}

func (c *biteshipClient) CreateOrder(req BiteshipCreateOrderRequest) (BiteshipOrderResult, error) {
	raw, err := c.do("POST", "/v1/orders", req)
	if err != nil {
		return BiteshipOrderResult{}, err
	}

	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
		Id      string `json:"id"`
		Courier *struct {
			TrackingId string `json:"tracking_id"`
			WaybillId  string `json:"waybill_id"`
		} `json:"courier"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return BiteshipOrderResult{}, err
	}
	if !resp.Success || resp.Courier == nil {
		return BiteshipOrderResult{}, fmt.Errorf("biteship CreateOrder failed: %s", resp.Error)
	}
	return BiteshipOrderResult{
		OrderId:    resp.Id,
		WaybillId:  resp.Courier.WaybillId,
		TrackingId: resp.Courier.TrackingId,
	}, nil
}

func (c *biteshipClient) GetTracking(trackingId string) (map[string]interface{}, error) {
	path := "/v1/trackings/" + url.PathEscape(trackingId)
	raw, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}
