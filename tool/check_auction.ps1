Write-Host "Waiting 15s for auction worker to close auction..."
Start-Sleep -Seconds 15
$r = Invoke-RestMethod "http://localhost:8080/auctions/e6653488-8e40-4142-a503-0a1e582176ac" -Method GET
$r | ConvertTo-Json -Depth 5
