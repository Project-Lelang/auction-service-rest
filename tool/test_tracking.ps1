$BASE = "http://localhost:8080"
$auctionId  = "e6653488-8e40-4142-a503-0a1e582176ac"
$shipmentId = "5f6f499b-09e5-4dc3-9442-2995e7845106"

$buyerToken  = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62222222222","password":"Buyer@123"}').data.token
$sellerToken = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62111111111","password":"Seller@123"}').data.token

# Check auction status
Write-Host "[1] Auction status"
$auc = Invoke-RestMethod "$BASE/auctions/$auctionId" -Method GET -Headers @{Authorization="Bearer $sellerToken"}
Write-Host "  Status: $($auc.data.auction.status)"

# Get tracking
Write-Host "`n[2] Get tracking"
try {
    $t = Invoke-RestMethod "$BASE/auctions/$auctionId/shipments/$shipmentId/tracking" -Method GET -Headers @{Authorization="Bearer $sellerToken"}
    $t | ConvertTo-Json -Depth 8
} catch {
    Write-Host "  ERROR (seller): $($_.ErrorDetails.Message)"
}

# Try as buyer
Write-Host "`n[3] Get tracking as buyer"
try {
    $t2 = Invoke-RestMethod "$BASE/auctions/$auctionId/shipments/$shipmentId/tracking" -Method GET -Headers @{Authorization="Bearer $buyerToken"}
    $t2 | ConvertTo-Json -Depth 8
} catch {
    Write-Host "  ERROR (buyer): $($_.ErrorDetails.Message)"
}
