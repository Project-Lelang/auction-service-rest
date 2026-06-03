#!/usr/bin/env powershell
# Continue from: product already VERIFIED, addresses already exist

$BASE = "http://localhost:8080"
$ErrorActionPreference = "Continue"

# Known IDs from DB
$productId    = "e45c1752-1d14-4a05-b9de-fc6b81fbd5ee"
$sellerAddrId = "6c1e29ee-5ed6-4800-8e9a-5f65e51b5583"
$buyerAddrId  = "96971fb8-b376-40fe-baf7-84cb0ee074bd"

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
Step 1 "Login"
$adminToken  = (Invoke-Api POST "/admin/auth/login" @{phone="+62000000";password="SuperAdmin@123"} $null).data.token
$sellerToken = (Invoke-Api POST "/auth/login" @{phone="+62111111111";password="Seller@123"} $null).data.token
$buyerToken  = (Invoke-Api POST "/auth/login" @{phone="+62222222222";password="Buyer@123"} $null).data.token
if (!$sellerToken) { FAIL "Seller login" }
if (!$buyerToken)  { FAIL "Buyer login" }
OK "All tokens acquired"

# Create auction
Step 2 "Seller creates auction"
$startAt = (Get-Date).AddSeconds(5).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$endAt   = (Get-Date).AddSeconds(50).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$auction = Invoke-Api POST "/own/auctions" @{
    product_id     = $productId
    starting_price = 500000
    start_time     = $startAt
    end_time       = $endAt
} $sellerToken
if (!$auction) { FAIL "Create auction failed" }
$auctionId = $auction.data.auction.id
OK "Auction created: $auctionId (starts in 5s, ends in 50s)"

Write-Host "  Waiting 8s for auction to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 8

# Buyer places bid
Step 3 "Buyer places bid"
$bid = Invoke-Api POST "/auctions/$auctionId/bids" @{amount=600000} $buyerToken
if (!$bid) { FAIL "Place bid failed" }
$bidId = $bid.data.auction_bid.id
OK "Bid placed: $bidId (amount: 600000)"

# Wait for auction to end
Write-Host "  Waiting 45s for auction to end..." -ForegroundColor Yellow
Start-Sleep -Seconds 45

# Check auction status
Step 4 "Check auction status"
$auctionData = Invoke-Api GET "/auctions/$auctionId" $null $sellerToken
OK "Auction status: $($auctionData.data.auction.status)"

# Check as admin
$adminAuction = Invoke-Api GET "/admin/auctions/$auctionId" $null $adminToken
if ($adminAuction) {
    OK "Admin auction status: $($adminAuction.data.auction.status)"
}

Write-Host "`n=== Summary ===" -ForegroundColor Cyan
Write-Host "Auction ID   : $auctionId"
Write-Host "Bid ID       : $bidId"
Write-Host "Seller addr  : $sellerAddrId"
Write-Host "Buyer addr   : $buyerAddrId"

# Save state
@{
    productId    = $productId
    auctionId    = $auctionId
    bidId        = $bidId
    sellerAddrId = $sellerAddrId
    buyerAddrId  = $buyerAddrId
    sellerToken  = $sellerToken
    buyerToken   = $buyerToken
    adminToken   = $adminToken
} | ConvertTo-Json | Set-Content "D:\go\src\auction-service\tool\test_state.json"
OK "State saved to tool/test_state.json"
