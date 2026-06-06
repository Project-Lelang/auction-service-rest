$BASE = "http://localhost:8080"
$auctionId = "e6653488-8e40-4142-a503-0a1e582176ac"
$paymentId = "eabe7e5c-c7a0-4925-9bfa-675187488548"

$buyerToken = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62222222222","password":"Buyer@123"}').data.token
$sellerToken = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62111111111","password":"Seller@123"}').data.token
$adminToken = (Invoke-RestMethod "$BASE/admin/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62000000","password":"SuperAdmin@123"}').data.token
Write-Host "Tokens acquired"

# Get payment
Write-Host "`n[1] Get payment"
try {
    $payment = Invoke-RestMethod "$BASE/auctions/$auctionId/payments/$paymentId" -Method GET -Headers @{Authorization="Bearer $buyerToken"}
    $payment | ConvertTo-Json -Depth 5
} catch {
    Write-Host "  ERROR: $($_.ErrorDetails.Message)"
}

# Get auction winners
Write-Host "`n[2] Get auction winners"
try {
    $winners = Invoke-RestMethod "$BASE/auctions/$auctionId/winners/filter" -Method POST -ContentType "application/json" -Headers @{Authorization="Bearer $buyerToken"} -Body '{}'
    $winners | ConvertTo-Json -Depth 5
} catch {
    Write-Host "  ERROR: $($_.ErrorDetails.Message)"
}

# Get shipments
Write-Host "`n[3] Get shipments"
try {
    $shipments = Invoke-RestMethod "$BASE/auctions/$auctionId/shipments" -Method GET -Headers @{Authorization="Bearer $buyerToken"}
    $shipments | ConvertTo-Json -Depth 5
} catch {
    Write-Host "  ERROR: $($_.ErrorDetails.Message)"
}
