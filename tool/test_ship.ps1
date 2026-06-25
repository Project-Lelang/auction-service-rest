$BASE = "http://localhost:8080"
$auctionId  = "e6653488-8e40-4142-a503-0a1e582176ac"
$shipmentId = "5f6f499b-09e5-4dc3-9442-2995e7845106"
$buyerAddrId = "96971fb8-b376-40fe-baf7-84cb0ee074bd"

$buyerToken  = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62222222222","password":"Buyer@123"}').data.token
$sellerToken = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62111111111","password":"Seller@123"}').data.token
Write-Host "Tokens OK"

# Step 1: Buyer confirms address
Write-Host "`n[1] Buyer confirms address"
$confirmBody = @{address_id = $buyerAddrId} | ConvertTo-Json
try {
    $r = Invoke-RestMethod "$BASE/auctions/$auctionId/shipments/$shipmentId/buyer-address" -Method PATCH -ContentType "application/json" -Headers @{Authorization="Bearer $buyerToken"} -Body $confirmBody
    $r | ConvertTo-Json -Depth 3
    Write-Host "  Buyer address confirmed!"
} catch {
    Write-Host "  ERROR: $($_.ErrorDetails.Message)"
}

# Check auction status
$auc = Invoke-RestMethod "$BASE/auctions/$auctionId" -Method GET -Headers @{Authorization="Bearer $sellerToken"}
Write-Host "  Auction status: $($auc.data.auction.status)"

# Step 2: Seller ships (call Biteship CreateOrder!)
Write-Host "`n[2] Seller ships via Biteship"
$shipBody = @{courier_code="jne"; service_code="REG"} | ConvertTo-Json
try {
    $s = Invoke-RestMethod "$BASE/auctions/$auctionId/shipments/$shipmentId/ship" -Method POST -ContentType "application/json" -Headers @{Authorization="Bearer $sellerToken"} -Body $shipBody
    $s | ConvertTo-Json -Depth 8
    Write-Host "  SHIPPED!"
} catch {
    Write-Host "  ERROR: $($_.ErrorDetails.Message)"
}
