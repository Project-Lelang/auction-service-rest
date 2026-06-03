$BASE = "http://localhost:8080"
$auctionId  = "e6653488-8e40-4142-a503-0a1e582176ac"
$shipmentId = "5f6f499b-09e5-4dc3-9442-2995e7845106"

$sellerToken = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62111111111","password":"Seller@123"}').data.token
Write-Host "Seller token OK"

# Try different courier/service combinations
$combos = @(
    @{courier_code="jne"; service_code="reg"},
    @{courier_code="jne"; service_code="yes"},
    @{courier_code="jne"; service_code="oke"},
    @{courier_code="sicepat"; service_code="reg"},
    @{courier_code="jnt"; service_code="ez"}
)

foreach ($combo in $combos) {
    Write-Host "`nTrying: courier=$($combo.courier_code) service=$($combo.service_code)"
    $shipBody = $combo | ConvertTo-Json
    try {
        $s = Invoke-RestMethod "$BASE/auctions/$auctionId/shipments/$shipmentId/ship" -Method POST -ContentType "application/json" -Headers @{Authorization="Bearer $sellerToken"} -Body $shipBody
        Write-Host "SUCCESS!"
        $s | ConvertTo-Json -Depth 8
        break
    } catch {
        Write-Host "  ERROR: $($_.ErrorDetails.Message)"
    }
}
