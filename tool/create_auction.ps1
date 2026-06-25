$BASE = "http://localhost:8080"
$sellerToken = (Invoke-RestMethod "$BASE/auth/login" -Method POST -ContentType "application/json" -Body '{"phone":"+62111111111","password":"Seller@123"}').data.token
$startAt = (Get-Date).AddHours(1).AddMinutes(2).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$endAt   = (Get-Date).AddHours(2).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
Write-Host "start_at: $startAt  end_at: $endAt"
$body = @{product_id="e45c1752-1d14-4a05-b9de-fc6b81fbd5ee"; starting_price=500000; start_time=$startAt; end_time=$endAt} | ConvertTo-Json
$result = Invoke-RestMethod "$BASE/own/auctions" -Method POST -ContentType "application/json" -Headers @{Authorization="Bearer $sellerToken"} -Body $body
$result | ConvertTo-Json -Depth 5
