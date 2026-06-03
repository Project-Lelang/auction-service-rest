$BASE = "http://localhost:8080"
$auctionId = "e6653488-8e40-4142-a503-0a1e582176ac"
$paymentId = "eabe7e5c-c7a0-4925-9bfa-675187488548"
$winnerId  = "e06abdfe-246b-424e-8526-9d71d4774c0b"

$buyerToken  = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62222222222","password":"Buyer@123"}').data.token
$sellerToken = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62111111111","password":"Seller@123"}').data.token
$adminToken  = (Invoke-RestMethod "$BASE/admin/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62000000","password":"SuperAdmin@123"}').data.token
Write-Host "Tokens OK"

# Try winners as seller
Write-Host "`n[1] Winners (as seller)"
try {
    $w = Invoke-RestMethod "$BASE/auctions/$auctionId/winners/filter" -Method POST -ContentType "application/json" -Headers @{Authorization="Bearer $sellerToken"} -Body '{}'
    $w | ConvertTo-Json -Depth 5
} catch { Write-Host "  ERROR: $($_.ErrorDetails.Message)" }

# Try winners as admin
Write-Host "`n[2] Winners (as admin)"
try {
    $w2 = Invoke-RestMethod "$BASE/admin/auctions/$auctionId" -Method GET -Headers @{Authorization="Bearer $adminToken"}
    $w2 | ConvertTo-Json -Depth 5
} catch { Write-Host "  ERROR: $($_.ErrorDetails.Message)" }

# Shipments as seller
Write-Host "`n[3] Shipments (as seller)"
try {
    $s = Invoke-RestMethod "$BASE/auctions/$auctionId/shipments" -Method GET -Headers @{Authorization="Bearer $sellerToken"}
    $s | ConvertTo-Json -Depth 5
} catch { Write-Host "  ERROR: $($_.ErrorDetails.Message)" }
