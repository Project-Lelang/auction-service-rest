$conf = Get-Content D:\go\src\auction-service\conf.yml -Raw
$key = ([regex]'api_key:\s*(.+)').Match($conf).Groups[1].Value.Trim()
Write-Host "Using Biteship key: $($key.Substring(0,20))..."

$headers = @{
    'Authorization' = "Bearer $key"
    'Content-Type'  = 'application/json'
}
$body = @{
    origin_area_id      = 'IDNP6IDNC148IDND843IDZ12250'
    destination_area_id = 'IDNP6IDNC147IDND832IDZ10310'
    couriers            = 'jne,sicepat,jnt,anteraja,ninja,lion,sap'
    items               = @(@{name='Laptop';description='test';value=600000;length=10;width=10;height=10;weight=200;quantity=1})
} | ConvertTo-Json -Depth 5

try {
    $r = Invoke-RestMethod 'https://api.biteship.com/v1/rates/couriers' -Method POST -Headers $headers -Body $body
    $r.pricing | Select-Object courier_name, courier_code, courier_service_name, courier_service_code, price | Format-Table -AutoSize
} catch {
    Write-Host "ERROR: $($_.ErrorDetails.Message)"
}
