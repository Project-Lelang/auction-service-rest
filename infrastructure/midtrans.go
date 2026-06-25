package infrastructure

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MidtransSnapRequest is the request body for the Midtrans Snap API.
type MidtransSnapRequest struct {
	TransactionDetails MidtransTransactionDetails `json:"transaction_details"`
	ItemDetails        []MidtransItemDetail       `json:"item_details,omitempty"`
	CustomerDetails    MidtransCustomerDetails    `json:"customer_details"`
	Expiry             *MidtransExpiry            `json:"expiry,omitempty"`
}

// MidtransTransactionDetails holds order and amount details.
type MidtransTransactionDetails struct {
	OrderId     string  `json:"order_id"`
	GrossAmount float64 `json:"gross_amount"`
}

// MidtransItemDetail makes the Snap payload explicit instead of sending only
// the gross amount.
type MidtransItemDetail struct {
	Id       string  `json:"id"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Name     string  `json:"name"`
}

// MidtransCustomerDetails holds buyer information passed to Midtrans.
type MidtransCustomerDetails struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

// MidtransExpiry controls how long a Snap token stays valid.
type MidtransExpiry struct {
	StartTime string `json:"start_time"` // "yyyy-MM-dd HH:mm:ss Z"
	Unit      string `json:"unit"`       // minute | hour | day
	Duration  int    `json:"duration"`
}

// MidtransSnapResponse is returned by the Midtrans Snap create-transaction endpoint.
type MidtransSnapResponse struct {
	Token       string `json:"token"`
	RedirectUrl string `json:"redirect_url"`
}

// MidtransNotification is the webhook payload Midtrans sends on payment events.
type MidtransNotification struct {
	TransactionStatus string `json:"transaction_status"`
	OrderId           string `json:"order_id"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
	GrossAmount       string `json:"gross_amount"`
}

// MidtransClient abstracts calls to the Midtrans Snap API.
type MidtransClient interface {
	CreateSnapTransaction(request MidtransSnapRequest) (*MidtransSnapResponse, error)
}

type midtransClient struct {
	serverKey string
	baseUrl   string
}

// NewMidtransClient creates a Midtrans HTTP client.
// When isSandbox is true, the sandbox base URL is used.
func NewMidtransClient(serverKey string, isSandbox bool) MidtransClient {
	baseUrl := "https://app.midtrans.com"
	if isSandbox {
		baseUrl = "https://app.sandbox.midtrans.com"
	}
	return &midtransClient{serverKey: serverKey, baseUrl: baseUrl}
}

func (c *midtransClient) authHeader() string {
	encoded := base64.StdEncoding.EncodeToString([]byte(c.serverKey + ":"))
	return "Basic " + encoded
}

// CreateSnapTransaction calls POST /snap/v1/transactions and returns the token and
// redirect URL that the frontend uses to open the Snap payment page.
func (c *midtransClient) CreateSnapTransaction(request MidtransSnapRequest) (*MidtransSnapResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("midtrans: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/snap/v1/transactions", c.baseUrl)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("midtrans: build request: %w", err)
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("midtrans: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("midtrans: unexpected status %d: %v", resp.StatusCode, errBody)
	}

	var snapResp MidtransSnapResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapResp); err != nil {
		return nil, fmt.Errorf("midtrans: decode response: %w", err)
	}

	return &snapResp, nil
}
