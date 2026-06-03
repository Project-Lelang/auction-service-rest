$BASE = "http://localhost:8080"
$auctionId = "e6653488-8e40-4142-a503-0a1e582176ac"
$buyerToken = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62222222222","password":"Buyer@123"}').data.token
Write-Host "Buyer token: $($buyerToken.Substring(0,20))..."

# Place bid
try {
    $bid = Invoke-RestMethod "$BASE/auctions/$auctionId/bids" -Method POST -ContentType "application/json" -Headers @{Authorization="Bearer $buyerToken"} -Body '{"amount":600000}'
    Write-Host "BID OK: $($bid | ConvertTo-Json -Depth 5)"
} catch {
    Write-Host "BID ERROR: $($_.ErrorDetails.Message)"
}
