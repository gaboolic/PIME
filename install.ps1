# PIME 输入法安装脚本
# 复制运行时文件并注册/启动 PIME

param(
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"

$InstallDir = "C:\Program Files (x86)\PIME"
$SourceDir = Split-Path -Parent $MyInvocation.MyCommand.Path

$BackendsJsonSource = Join-Path $SourceDir "backends.json"
$PIMELauncherSource = Join-Path $SourceDir "build\PIMELauncher\Release\PIMELauncher.exe"
$X86DllSource = Join-Path $SourceDir "build\PIMETextService\Release\PIMETextService.dll"
$X64DllSource = Join-Path $SourceDir "build64\PIMETextService\Release\PIMETextService.dll"
$PythonSource = Join-Path $SourceDir "python"
$NodeSource = Join-Path $SourceDir "node"
$GoBackendBuildSource = Join-Path $SourceDir "go-backend\build\go-backend"
$GoBackendSource = Join-Path $SourceDir "go-backend"

$BackendsJsonDest = Join-Path $InstallDir "backends.json"
$PIMELauncherDest = Join-Path $InstallDir "PIMELauncher.exe"
$PythonDest = Join-Path $InstallDir "python"
$NodeDest = Join-Path $InstallDir "node"
$GoBackendDest = Join-Path $InstallDir "go-backend"

$X86Dir = Join-Path $InstallDir "x86"
$X64Dir = Join-Path $InstallDir "x64"
$X86DllDest = Join-Path $X86Dir "PIMETextService.dll"
$X64DllDest = Join-Path $X64Dir "PIMETextService.dll"

$Regsvr32X86 = Join-Path $env:WINDIR "SysWOW64\regsvr32.exe"
$Regsvr32X64 = Join-Path $env:WINDIR "System32\regsvr32.exe"
$RunKeyPath = "HKLM:\Software\Microsoft\Windows\CurrentVersion\Run"
$RunKeyName = "PIMELauncher"

function Test-Admin {
    return ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole(
        [Security.Principal.WindowsBuiltInRole]::Administrator
    )
}

function Assert-Admin {
    if (-not (Test-Admin)) {
        throw "需要以管理员身份运行此脚本！请右键点击 PowerShell 选择'以管理员身份运行'"
    }
}

function Assert-PathExists {
    param(
        [string]$Path,
        [string]$Message
    )

    if (-not (Test-Path -LiteralPath $Path)) {
        throw $Message
    }
}

function New-DirectoryIfMissing {
    param([string]$Path)

    New-Item -ItemType Directory -Path $Path -Force | Out-Null
}

function Copy-Directory {
    param(
        [string]$Source,
        [string]$Destination
    )

    Assert-PathExists -Path $Source -Message "找不到目录: $Source"

    if (Test-Path -LiteralPath $Destination) {
        Remove-Item -LiteralPath $Destination -Recurse -Force
    }

    New-Item -ItemType Directory -Path $Destination -Force | Out-Null
    Copy-Item -Path (Join-Path $Source "*") -Destination $Destination -Recurse -Force
}

function Invoke-Regsvr32 {
    param(
        [string]$Regsvr32Path,
        [string]$DllPath,
        [switch]$Unregister
    )

    $arguments = @()
    if ($Unregister) {
        $arguments += "/u"
    }
    $arguments += "/s"
    $arguments += $DllPath

    $process = Start-Process -FilePath $Regsvr32Path -ArgumentList $arguments -PassThru -Wait -WindowStyle Hidden
    if ($process.ExitCode -ne 0) {
        $action = if ($Unregister) { "注销" } else { "注册" }
        throw "${action}失败: $DllPath (退出码: $($process.ExitCode))"
    }
}

function Get-GoBackendInstallSource {
    if (Test-Path -LiteralPath $GoBackendBuildSource) {
        return $GoBackendBuildSource
    }

    return $GoBackendSource
}

function Stop-PIMELauncher {
    $launcherProcesses = Get-Process -Name "PIMELauncher" -ErrorAction SilentlyContinue
    if (-not $launcherProcesses) {
        return
    }

    Write-Host "停止正在运行的 PIMELauncher..." -ForegroundColor Yellow

    if (Test-Path -LiteralPath $PIMELauncherDest) {
        try {
            Start-Process -FilePath $PIMELauncherDest -ArgumentList "/quit" -WindowStyle Hidden
            Start-Sleep -Seconds 2
        } catch {
        }
    }

    $launcherProcesses = Get-Process -Name "PIMELauncher" -ErrorAction SilentlyContinue
    if ($launcherProcesses) {
        $launcherProcesses | Stop-Process -Force
        Start-Sleep -Seconds 1
    }

    Write-Host "✓ PIMELauncher 已停止" -ForegroundColor Green
}

function Set-PIMELauncherAutoStart {
    New-Item -Path $RunKeyPath -Force | Out-Null
    Set-ItemProperty -Path $RunKeyPath -Name $RunKeyName -Value $PIMELauncherDest
}

function Remove-PIMELauncherAutoStart {
    if (Get-ItemProperty -Path $RunKeyPath -Name $RunKeyName -ErrorAction SilentlyContinue) {
        Remove-ItemProperty -Path $RunKeyPath -Name $RunKeyName
    }
}

function Start-PIMELauncher {
    Assert-PathExists -Path $PIMELauncherDest -Message "找不到 PIMELauncher.exe: $PIMELauncherDest"
    Start-Process -FilePath $PIMELauncherDest
}

function Install-PIME {
    Write-Host "=== PIME 输入法安装 ===" -ForegroundColor Cyan

    Assert-Admin

    $ResolvedGoBackendSource = Get-GoBackendInstallSource

    Assert-PathExists -Path $BackendsJsonSource -Message "找不到 backends.json: $BackendsJsonSource"
    Assert-PathExists -Path $PIMELauncherSource -Message "找不到 PIMELauncher.exe: $PIMELauncherSource`n请先编译 PIMELauncher"
    Assert-PathExists -Path $X86DllSource -Message "找不到 32位 DLL: $X86DllSource`n请先编译 Win32 版本"
    Assert-PathExists -Path $X64DllSource -Message "找不到 64位 DLL: $X64DllSource`n请先编译 x64 版本"
    Assert-PathExists -Path $PythonSource -Message "找不到 python 目录: $PythonSource"
    Assert-PathExists -Path $NodeSource -Message "找不到 node 目录: $NodeSource"
    Assert-PathExists -Path $ResolvedGoBackendSource -Message "找不到 go-backend 目录: $ResolvedGoBackendSource`n请先运行 go-backend\build.bat 或准备好 go-backend 运行目录"

    Stop-PIMELauncher

    Write-Host "创建安装目录..." -ForegroundColor Yellow
    New-DirectoryIfMissing -Path $InstallDir
    New-DirectoryIfMissing -Path $X86Dir
    New-DirectoryIfMissing -Path $X64Dir
    Write-Host "✓ 目录创建完成" -ForegroundColor Green

    Write-Host "`n复制 backends.json..." -ForegroundColor Yellow
    Copy-Item -LiteralPath $BackendsJsonSource -Destination $BackendsJsonDest -Force
    Write-Host "✓ backends.json 复制完成" -ForegroundColor Green

    Write-Host "`n复制 PIMELauncher.exe..." -ForegroundColor Yellow
    Copy-Item -LiteralPath $PIMELauncherSource -Destination $PIMELauncherDest -Force
    Write-Host "✓ PIMELauncher.exe 复制完成" -ForegroundColor Green

    Write-Host "`n复制 DLL 文件..." -ForegroundColor Yellow
    Copy-Item -LiteralPath $X86DllSource -Destination $X86DllDest -Force
    Copy-Item -LiteralPath $X64DllSource -Destination $X64DllDest -Force
    Write-Host "✓ DLL 复制完成" -ForegroundColor Green

    Write-Host "`n复制 python 目录..." -ForegroundColor Yellow
    Copy-Directory -Source $PythonSource -Destination $PythonDest
    Write-Host "✓ python 复制完成" -ForegroundColor Green

    Write-Host "`n复制 node 目录..." -ForegroundColor Yellow
    Copy-Directory -Source $NodeSource -Destination $NodeDest
    Write-Host "✓ node 复制完成" -ForegroundColor Green

    Write-Host "`n复制 go-backend 目录..." -ForegroundColor Yellow
    Copy-Directory -Source $ResolvedGoBackendSource -Destination $GoBackendDest
    Write-Host "✓ go-backend 复制完成" -ForegroundColor Green

    Write-Host "`n注册 PIME 输入法服务..." -ForegroundColor Yellow
    Invoke-Regsvr32 -Regsvr32Path $Regsvr32X86 -DllPath $X86DllDest
    Write-Host "✓ 32位服务注册成功" -ForegroundColor Green
    Invoke-Regsvr32 -Regsvr32Path $Regsvr32X64 -DllPath $X64DllDest
    Write-Host "✓ 64位服务注册成功" -ForegroundColor Green

    Write-Host "`n写入开机自启动..." -ForegroundColor Yellow
    Set-PIMELauncherAutoStart
    Write-Host "✓ 已写入 PIMELauncher 自启动" -ForegroundColor Green

    Write-Host "`n启动 PIMELauncher..." -ForegroundColor Yellow
    Start-PIMELauncher
    Write-Host "✓ PIMELauncher 已启动" -ForegroundColor Green

    Write-Host "`n=== 安装完成 ===" -ForegroundColor Cyan
    Write-Host "PIME 输入法已成功安装到: $InstallDir" -ForegroundColor Green
    Write-Host "如果输入法列表没有立即刷新，请注销重新登录或重启系统。" -ForegroundColor Yellow
}

function Uninstall-PIME {
    Write-Host "=== PIME 输入法卸载 ===" -ForegroundColor Cyan

    Assert-Admin
    Stop-PIMELauncher

    if (Test-Path -LiteralPath $X86DllDest) {
        Write-Host "注销 32位 DLL..." -ForegroundColor Yellow
        Invoke-Regsvr32 -Regsvr32Path $Regsvr32X86 -DllPath $X86DllDest -Unregister
        Write-Host "✓ 32位服务已注销" -ForegroundColor Green
    }

    if (Test-Path -LiteralPath $X64DllDest) {
        Write-Host "注销 64位 DLL..." -ForegroundColor Yellow
        Invoke-Regsvr32 -Regsvr32Path $Regsvr32X64 -DllPath $X64DllDest -Unregister
        Write-Host "✓ 64位服务已注销" -ForegroundColor Green
    }

    Write-Host "移除 PIMELauncher 自启动..." -ForegroundColor Yellow
    Remove-PIMELauncherAutoStart
    Write-Host "✓ 已移除 PIMELauncher 自启动" -ForegroundColor Green

    if (Test-Path -LiteralPath $InstallDir) {
        Write-Host "删除安装目录..." -ForegroundColor Yellow
        Remove-Item -LiteralPath $InstallDir -Recurse -Force
        Write-Host "✓ 安装目录已删除" -ForegroundColor Green
    }

    Write-Host "`n=== 卸载完成 ===" -ForegroundColor Cyan
    Write-Host "PIME 输入法已从系统中移除" -ForegroundColor Green
}

if ($Uninstall) {
    Uninstall-PIME
} else {
    Install-PIME
}
