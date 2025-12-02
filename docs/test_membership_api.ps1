# 会员系统 API 测试脚本
# 使用方法: .\test_membership_api.ps1

$baseUrl = "http://localhost:8096"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "会员系统 API 测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. 获取所有会员等级
Write-Host "1. 获取所有会员等级" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/tiers" -Method GET
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 2. 免费用户 - 获取会员信息
Write-Host "2. 免费用户 (ID=1) - 获取会员信息" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/info" -Method GET -Headers @{"X-User-ID" = "1" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 3. 基础用户 - 获取会员信息
Write-Host "3. 基础用户 (ID=2) - 获取会员信息" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/info" -Method GET -Headers @{"X-User-ID" = "2" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 4. 专业用户 - 获取会员信息
Write-Host "4. 专业用户 (ID=3) - 获取会员信息" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/info" -Method GET -Headers @{"X-User-ID" = "3" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 5. 企业用户 - 获取会员信息
Write-Host "5. 企业用户 (ID=4) - 获取会员信息" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/info" -Method GET -Headers @{"X-User-ID" = "4" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 6. 免费用户 - 获取配额信息
Write-Host "6. 免费用户 - 获取配额信息" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/quota" -Method GET -Headers @{"X-User-ID" = "1" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 7. 免费用户 - 获取可用功能
Write-Host "7. 免费用户 - 获取可用功能" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/features" -Method GET -Headers @{"X-User-ID" = "1" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 8. 专业用户 - 获取可用功能
Write-Host "8. 专业用户 - 获取可用功能" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/features" -Method GET -Headers @{"X-User-ID" = "3" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 9. 免费用户 - 检查AI翻译功能
Write-Host "9. 免费用户 - 检查AI翻译功能 (应该不允许)" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/features/ai_translation/check" -Method GET -Headers @{"X-User-ID" = "1" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 10. 基础用户 - 检查AI翻译功能
Write-Host "10. 基础用户 - 检查AI翻译功能 (应该允许)" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/features/ai_translation/check" -Method GET -Headers @{"X-User-ID" = "2" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 11. 免费用户 - 检查API访问功能
Write-Host "11. 免费用户 - 检查API访问功能 (应该不允许)" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/features/api_access/check" -Method GET -Headers @{"X-User-ID" = "1" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 12. 企业用户 - 检查API访问功能
Write-Host "12. 企业用户 - 检查API访问功能 (应该允许)" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/features/api_access/check" -Method GET -Headers @{"X-User-ID" = "4" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

# 13. 获取加油包状态
Write-Host "13. 免费用户 - 获取加油包状态" -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$baseUrl/api/v1/membership/boost-pack" -Method GET -Headers @{"X-User-ID" = "1" }
$response | ConvertTo-Json -Depth 10
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "测试完成!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
