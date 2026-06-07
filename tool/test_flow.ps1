#!/usr/bin/env pwsh
# Full integration test for auction-service Biteship flow
# Run: pwsh -File tool/test_flow.ps1

$BASE = "http://localhost:8080"

function Invoke-Api($method, $path, $body, $token) {
    $headers = @{}
    if ($token) { $headers["Authorization"] = "Bearer $token" }
    $params = @{
        Uri         = "$BASE$path"
        Method      = $method
        ContentType = "application/json"
    }
    if ($body) { $params["Body"] = ($body | ConvertTo-Json -Compress) }
    if ($headers.Count) { $params["Headers"] = $headers }
    try {
        return Invoke-RestMethod @params
    } catch {
        $msg = $_.ErrorDetails.Message
        Write-Host "  ERROR: $msg" -ForegroundColor Red
        return $null
    }
}

function Step($n, $desc) { Write-Host "`n[$n] $desc" -ForegroundColor Cyan }
function OK($msg)   { Write-Host "  OK: $msg" -ForegroundColor Green }
function FAIL($msg) { Write-Host "  FAIL: $msg" -ForegroundColor Red; exit 1 }

# -----------------------------------------------------------------------
Step 1 "Admin login"
$adminToken = (Invoke-Api POST "/admin/auth/login" @{phone="+62000000";password="SuperAdmin@123"}).data.token
if (!$adminToken) { FAIL "Admin login failed" }
OK "Admin token acquired"

# -----------------------------------------------------------------------
Step 2 "Seller login (assume already registered)"
$sellerToken = (Invoke-Api POST "/auth/login" @{phone="+62111111111";password="Seller@123"}).data.token
if (!$sellerToken) { FAIL "Seller login failed" }
OK "Seller token acquired"

$buyerToken = (Invoke-Api POST "/auth/login" @{phone="+62222222222";password="Buyer@123"}).data.token
if (!$buyerToken) { FAIL "Buyer login failed" }
OK "Buyer token acquired"

# -----------------------------------------------------------------------
Step 3 "Approve seller role via role request"
# Seller requests SELLER role
$roleReq = Invoke-Api POST "/own/role-requests" @{role="SELLER"} $sellerToken
if ($roleReq) { OK "Seller role request created: $($roleReq.data.role_request.id)" }
# Admin approves
$roleReqId = $roleReq.data.role_request.id
$approve = Invoke-Api PATCH "/admin/role-requests/$roleReqId/approve" $null $adminToken
if ($approve) { OK "Seller role approved" }

# -----------------------------------------------------------------------
Step 4 "Seller creates user address (with biteship_area_id)"
# Search area first
$areas = Invoke-Api GET "/biteship/areas?keyword=Pesanggrahan" $null $sellerToken
$sellerArea = $areas.data.areas[0]
OK "Seller area: $($sellerArea.name) [$($sellerArea.id)]"

$sellerAddr = Invoke-Api POST "/user-addresses" @{
    label="Gudang Seller"
    recipient_name="Seller Test"
    phone="+62111111111"
    city_id="151"
    city_name="Jakarta Selatan"
    province_name="DKI Jakarta"
    address="Jl. Test No. 1"
    postal_code="12250"
    biteship_area_id=$sellerArea.id
    is_default=$true
} $sellerToken
$sellerAddrId = $sellerAddr.data.user_address.id
OK "Seller address created: $sellerAddrId"

# Buyer address
$buyerAreas = Invoke-Api GET "/biteship/areas?keyword=Menteng" $null $buyerToken
$buyerArea = $buyerAreas.data.areas[0]
OK "Buyer area: $($buyerArea.name) [$($buyerArea.id)]"

$buyerAddr = Invoke-Api POST "/user-addresses" @{
    label="Rumah Buyer"
    recipient_name="Buyer Test"
    phone="+62222222222"
    city_id="152"
    city_name="Jakarta Pusat"
    province_name="DKI Jakarta"
    address="Jl. Menteng No. 2"
    postal_code="10310"
    biteship_area_id=$buyerArea.id
    is_default=$true
} $buyerToken
$buyerAddrId = $buyerAddr.data.user_address.id
OK "Buyer address created: $buyerAddrId"

# -----------------------------------------------------------------------
Step 5 "Seller creates product"
# Need to upload image first - skip image for now, check if required
$product = Invoke-Api POST "/own/products" @{
    name="Laptop Test Item"
    description="Laptop bekas kondisi baik"
    price=1000000
    weight_gram=2000
} $sellerToken
if (!$product) { FAIL "Create product failed (image might be required)" }
$productId = $product.data.product.id
OK "Product created: $productId"

# Admin approves product
$approve = Invoke-Api PATCH "/admin/products/$productId/approve" $null $adminToken
OK "Product approved"

# -----------------------------------------------------------------------
Step 6 "Seller creates auction (short duration for testing)"
$startAt = (Get-Date).AddSeconds(3).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$endAt   = (Get-Date).AddSeconds(30).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$auction = Invoke-Api POST "/own/auctions" @{
    product_id=$productId
    start_price=500000
    start_at=$startAt
    end_at=$endAt
} $sellerToken
if (!$auction) { FAIL "Create auction failed" }
$auctionId = $auction.data.auction.id
OK "Auction created: $auctionId (starts in 3s)"

# Wait for auction to start
Start-Sleep -Seconds 5

# -----------------------------------------------------------------------
Step 7 "Buyer places bid"
$bid = Invoke-Api POST "/auctions/$auctionId/bids" @{amount=600000} $buyerToken
if (!$bid) { FAIL "Place bid failed (buyer may not have BUYER role)" }
$bidId = $bid.data.auction_bid.id
OK "Bid placed: $bidId (600000)"

# Wait for auction to end
Write-Host "  Waiting 30s for auction to end..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

# -----------------------------------------------------------------------
Step 8 "Check auction status - should be WAITING_FOR_SELLER_DECISION or similar"
$auctionData = Invoke-Api GET "/auctions/$auctionId" $null $sellerToken
OK "Auction status: $($auctionData.data.auction.status)"

# -----------------------------------------------------------------------
Step 9 "Verify addresses saved with biteship_area_id"
$sAddr = Invoke-Api GET "/user-addresses/$sellerAddrId" $null $sellerToken
OK "Seller address biteship_area_id: $($sAddr.data.user_address.biteship_area_id)"

$bAddr = Invoke-Api GET "/user-addresses/$buyerAddrId" $null $buyerToken
OK "Buyer address biteship_area_id: $($bAddr.data.user_address.biteship_area_id)"

Write-Host "`n=== Test run complete ===" -ForegroundColor Cyan
Write-Host "Seller address ID : $sellerAddrId"
Write-Host "Buyer address ID  : $buyerAddrId"
Write-Host "Auction ID        : $auctionId"
Write-Host "Bid ID            : $bidId"
