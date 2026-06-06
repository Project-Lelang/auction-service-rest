$BASE = "http://localhost:8080"
$auctionId = "e6653488-8e40-4142-a503-0a1e582176ac"
$paymentId = "eabe7e5c-c7a0-4925-9bfa-675187488548"

$buyerToken  = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62222222222","password":"Buyer@123"}').data.token
$sellerToken = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62111111111","password":"Seller@123"}').data.token
Write-Host "Tokens OK"

# Simulate Midtrans payment notification (settlement)
Write-Host "`n[1] Simulate payment settlement webhook"
$notification = @{
    transaction_status = "settlement"
    order_id           = $paymentId
    fraud_status       = "accept"
    payment_type       = "bank_transfer"
    gross_amount       = "605000.00"
} | ConvertTo-Json

try {
    $result = Invoke-RestMethod "$BASE/payment-notifications" -Method POST -ContentType "application/json" -Body $notification
    $result | ConvertTo-Json -Depth 3
    Write-Host "  Payment notification sent!"
} catch {
    Write-Host "  ERROR: $($_.ErrorDetails.Message)"
}

Start-Sleep -Seconds 3

# Check auction status
Write-Host "`n[2] Check auction status"
try {
    $auction = Invoke-RestMethod "$BASE/auctions/$auctionId" -Method GET -Headers @{Authorization="Bearer $sellerToken"}
    Write-Host "  Auction status: $($auction.data.auction.status)"
} catch {
    Write-Host "  ERROR: $($_.ErrorDetails.Message)"
}

# Check shipments
Write-Host "`n[3] Check shipments (seller)"
try {
    $shipments = Invoke-RestMethod "$BASE/auctions/$auctionId/shipments" -Method GET -Headers @{Authorization="Bearer $sellerToken"}
    $shipments | ConvertTo-Json -Depth 8
} catch {
    Write-Host "  ERROR: $($_.ErrorDetails.Message)"
}

# Check shipments as buyer
Write-Host "`n[4] Check shipments (buyer)"
try {
    $shipments2 = Invoke-RestMethod "$BASE/auctions/$auctionId/shipments" -Method GET -Headers @{Authorization="Bearer $buyerToken"}
    $shipments2 | ConvertTo-Json -Depth 8
} catch {
    Write-Host "  ERROR: $($_.ErrorDetails.Message)"
}
