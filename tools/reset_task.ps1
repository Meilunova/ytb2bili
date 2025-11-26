# 任务重置工具
param(
    [Parameter(Mandatory = $true)]
    [string]$VideoId,
    
    [switch]$Clean
)

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "🔄 任务重置工具" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""

if ($Clean) {
    Write-Host "⚠️  将清理视频文件（保留字幕）" -ForegroundColor Yellow
    $arguments = @($VideoId, "clean")
}
else {
    $arguments = @($VideoId)
}

# 运行 Go 脚本
go run reset_task.go @arguments
