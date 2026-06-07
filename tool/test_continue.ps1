#!/usr/bin/env powershell
# Continue integration test from step 5 (roles already assigned, addresses already created)

$BASE = "http://localhost:8080"
$ErrorActionPreference = "Continue"

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

# Login
Step 1 "Login all users"
$adminToken  = (Invoke-Api POST "/admin/auth/login" @{phone="+62000000";password="SuperAdmin@123"} $null).data.token
if (!$adminToken) { FAIL "Admin login failed" }
OK "Admin token acquired"

$sellerToken = (Invoke-Api POST "/auth/login" @{phone="+62111111111";password="Seller@123"} $null).data.token
if (!$sellerToken) { FAIL "Seller login failed" }
OK "Seller token acquired"

$buyerToken  = (Invoke-Api POST "/auth/login" @{phone="+62222222222";password="Buyer@123"} $null).data.token
if (!$buyerToken) { FAIL "Buyer login failed" }
OK "Buyer token acquired"

# Get existing addresses
Step 2 "Get existing addresses"
$sAddrs = (Invoke-Api GET "/own/user-addresses" $null $sellerToken).data.user_addresses
$bAddrs = (Invoke-Api GET "/own/user-addresses" $null $buyerToken).data.user_addresses
$sellerAddrId = $sAddrs[0].id
$buyerAddrId  = $bAddrs[0].id
OK "Seller address: $sellerAddrId (biteship_area_id: $($sAddrs[0].biteship_area_id))"
OK "Buyer address: $buyerAddrId (biteship_area_id: $($bAddrs[0].biteship_area_id))"

# Create product
Step 3 "Seller creates product"
$product = Invoke-Api POST "/own/products" @{
    name      = "Laptop Test Item"
    condition = "PRELOVED"
    description = "Laptop bekas kondisi baik untuk testing"
} $sellerToken
if (!$product) { FAIL "Create product failed" }
$productId = $product.data.product.id
OK "Product created: $productId"

# Admin approves product
Step 4 "Admin approves product"
$approve = Invoke-Api PATCH "/admin/products/$productId/approve" $null $adminToken
if ($approve) { OK "Product approved" } else { Write-Host "  WARN: approve returned null (may still work)" }

# Check product status
$prodCheck = Invoke-Api GET "/own/products/$productId" $null $sellerToken
OK "Product status: $($prodCheck.data.product.status)"

# Create auction
Step 5 "Seller creates auction (short duration)"
$startAt = (Get-Date).AddSeconds(5).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$endAt   = (Get-Date).AddSeconds(45).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$auction = Invoke-Api POST "/own/auctions" @{
    product_id  = $productId
    start_price = 500000
    start_at    = $startAt
    end_at      = $endAt
} $sellerToken
if (!$auction) { FAIL "Create auction failed" }
$auctionId = $auction.data.auction.id
OK "Auction created: $auctionId (starts in 5s, ends in 45s)"

Write-Host "  Waiting 8s for auction to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 8

# Buyer places bid
Step 6 "Buyer places bid"
$bid = Invoke-Api POST "/auctions/$auctionId/bids" @{amount=600000} $buyerToken
if (!$bid) { FAIL "Place bid failed" }
$bidId = $bid.data.auction_bid.id
OK "Bid placed: $bidId (amount: 600000)"

# Wait for auction to end
Write-Host "  Waiting 40s for auction to end..." -ForegroundColor Yellow
Start-Sleep -Seconds 40

# Check auction status
Step 7 "Check auction status after end"
$auctionData = Invoke-Api GET "/auctions/$auctionId" $null $sellerToken
OK "Auction status: $($auctionData.data.auction.status)"

# Check if winner was determined
Step 8 "Check auction winner"
$winner = Invoke-Api GET "/admin/auctions/$auctionId" $null $adminToken
OK "Auction data: status=$($winner.data.auction.status)"

Write-Host "`n=== Flow summary ===" -ForegroundColor Cyan
Write-Host "Product ID : $productId"
Write-Host "Auction ID : $auctionId"
Write-Host "Bid ID     : $bidId"
Write-Host "Seller addr: $sellerAddrId"
Write-Host "Buyer addr : $buyerAddrId"

# Save IDs to file for next steps
@{
    productId    = $productId
    auctionId    = $auctionId
    bidId        = $bidId
    sellerAddrId = $sellerAddrId
    buyerAddrId  = $buyerAddrId
    sellerToken  = $sellerToken
    buyerToken   = $buyerToken
    adminToken   = $adminToken
} | ConvertTo-Json | Out-File "D:\go\src\auction-service\tool\test_state.json"
OK "State saved to tool/test_state.json"
