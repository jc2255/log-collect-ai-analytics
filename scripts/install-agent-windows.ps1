# ==========================================================
# LCA Agent - Windows 一键部署脚本 (PowerShell)
# 将 logcollect_win.exe 注册为 Windows 服务，异常退出自动重启
#
# 使用方式（以管理员身份运行 PowerShell）:
#   .\install-agent-windows.ps1
# ==========================================================

#Requires -RunAsAdministrator

# ---------- 配置区域（请根据实际修改） ----------
$AgentID       = if ($env:AGENT_ID) { $env:AGENT_ID } else { "agent-win-001" }
$AdminURL      = if ($env:ADMIN_URL) { $env:ADMIN_URL } else { "http://192.168.1.100:8080" }
$ApiURL        = if ($env:API_URL) { $env:API_URL } else { "http://192.168.1.100:8086" }
$BatchSize     = if ($env:BATCH_SIZE) { $env:BATCH_SIZE } else { "50" }
$FlushSeconds  = if ($env:FLUSH_SECONDS) { $env:FLUSH_SECONDS } else { "5" }
$PushConcurrency = if ($env:PUSH_CONCURRENCY) { $env:PUSH_CONCURRENCY } else { "5" }
$HostnameOverride = if ($env:HOSTNAME_OVERRIDE) { $env:HOSTNAME_OVERRIDE } else { $env:COMPUTERNAME }

# ---------- 安装路径 ----------
$InstallDir  = "C:\lca-agent"
$ServiceName = "LCAAgent"
$DisplayName = "LCA Log Collection Agent"
$BinaryName  = "logcollect_win.exe"
$ConfigFile  = "$InstallDir\config.yaml"

Write-Host "========================================"
Write-Host "  LCA Agent - Windows 部署"
Write-Host "========================================"
Write-Host "  Agent ID : $AgentID"
Write-Host "  Admin    : $AdminURL"
Write-Host "  API      : $ApiURL"
Write-Host "  Hostname : $HostnameOverride"
Write-Host "  安装目录 : $InstallDir"
Write-Host "========================================"

# ---------- 1. 停止已有服务 ----------
$existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($existingService) {
    Write-Host ">> 停止并删除已有服务..."
    Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
    sc.exe delete $ServiceName | Out-Null
    Start-Sleep -Seconds 1
}

# ---------- 2. 创建安装目录 ----------
New-Item -ItemType Directory -Path "$InstallDir\logs" -Force | Out-Null
New-Item -ItemType Directory -Path "$InstallDir\dead_letters" -Force | Out-Null

# ---------- 3. 复制二进制文件 ----------
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$BinarySrc = Join-Path $ScriptDir $BinaryName

if (-not (Test-Path $BinarySrc)) {
    $BinarySrc = Join-Path $ScriptDir "..\release\bin\$BinaryName"
}

if (-not (Test-Path $BinarySrc)) {
    Write-Host "ERROR: 未找到 $BinaryName 二进制文件" -ForegroundColor Red
    Write-Host "  请将编译好的 logcollect_win.exe 放到脚本所在目录或 release\bin\ 下"
    exit 1
}

Copy-Item $BinarySrc "$InstallDir\$BinaryName" -Force
Write-Host ">> 已安装二进制文件到 $InstallDir\$BinaryName"

# ---------- 4. 生成配置文件 ----------
$configContent = @"
# LCA Agent 配置文件（由 install-agent-windows.ps1 生成）
api_server: "$ApiURL"
admin_server: "$AdminURL"
agent_id: "$AgentID"
batch_size: $BatchSize
flush_seconds: $FlushSeconds
push_concurrency: $PushConcurrency
hostname: "$HostnameOverride"
"@

Set-Content -Path $ConfigFile -Value $configContent -Encoding UTF8
Write-Host ">> 已生成配置文件 $ConfigFile"

# ---------- 5. 创建 offsets.json ----------
$offsetsFile = "$InstallDir\offsets.json"
if (-not (Test-Path $offsetsFile)) {
    Set-Content -Path $offsetsFile -Value "[]" -Encoding UTF8
}

# ---------- 6. 注册 Windows 服务 ----------
$binPath = "`"$InstallDir\$BinaryName`" -config `"$ConfigFile`""

# 使用 sc.exe 创建服务
sc.exe create $ServiceName binPath= $binPath start= auto DisplayName= $DisplayName | Out-Null
sc.exe description $ServiceName "LCA Log Collection Agent - 日志采集与运维管理Agent" | Out-Null

# 配置服务失败自动重启策略：第1次失败等5秒重启，第2次等10秒，后续等30秒
sc.exe failure $ServiceName reset= 86400 actions= restart/5000/restart/10000/restart/30000 | Out-Null

Write-Host ">> 已创建 Windows 服务: $ServiceName"

# ---------- 7. 启动服务 ----------
Start-Service -Name $ServiceName
Start-Sleep -Seconds 2

$svc = Get-Service -Name $ServiceName
if ($svc.Status -eq "Running") {
    Write-Host ""
    Write-Host "======== 部署完成 ========" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "======== 服务启动异常，请检查日志 ========" -ForegroundColor Yellow
}

Write-Host "  查看状态: Get-Service $ServiceName"
Write-Host "  查看日志: Get-Content $InstallDir\logs\agent.log -Tail 50 -Wait"
Write-Host "  停止服务: Stop-Service $ServiceName"
Write-Host "  启动服务: Start-Service $ServiceName"
Write-Host "  卸载服务: Stop-Service $ServiceName; sc.exe delete $ServiceName"
Write-Host ""
Write-Host "  配置文件: $ConfigFile"
Write-Host "  安装目录: $InstallDir"
Write-Host "==============================="
