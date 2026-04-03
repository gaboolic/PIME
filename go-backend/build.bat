@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ============================================
echo  PIME Go 后端构建脚本
echo ============================================
echo.

REM 检查 Go 环境
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未找到 Go 环境，请先安装 Go
    echo 下载地址: https://golang.org/dl/
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do (
    echo [信息] Go 版本: %%i
)

REM 设置构建参数
set BUILD_DIR=build
set EXAMPLE_NAME=simple_ime
set EXAMPLE_DIR=example\%EXAMPLE_NAME%

echo.
echo ============================================
echo 步骤 1: 准备构建目录
echo ============================================
echo.

if not exist %BUILD_DIR% (
    mkdir %BUILD_DIR%
    echo [信息] 创建构建目录: %BUILD_DIR%
) else (
    echo [信息] 构建目录已存在: %BUILD_DIR%
    del /q %BUILD_DIR%\* >nul 2>nul
)

echo.
echo ============================================
echo 步骤 2: 下载依赖
echo ============================================
echo.

cd /d "%~dp0"
go mod tidy
if %errorlevel% neq 0 (
    echo [警告] go mod tidy 执行失败，继续构建...
)

echo.
echo ============================================
echo 步骤 3: 构建示例程序
echo ============================================
echo.

echo [信息] 构建 %EXAMPLE_NAME% ...

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

go build -ldflags="-s -w" -o %BUILD_DIR%\%EXAMPLE_NAME%.exe %EXAMPLE_DIR%

if %errorlevel% neq 0 (
    echo [错误] 构建失败！
    exit /b 1
)

echo [成功] 构建完成: %BUILD_DIR%\%EXAMPLE_NAME%.exe

echo.
echo ============================================
echo 步骤 4: 复制配置文件
echo ============================================
echo.

if exist %EXAMPLE_DIR%\ime.json (
    copy %EXAMPLE_DIR%\ime.json %BUILD_DIR% >nul
    echo [信息] 复制 ime.json
)

if exist %EXAMPLE_DIR%\icon.ico (
    copy %EXAMPLE_DIR%\icon.ico %BUILD_DIR% >nul
    echo [信息] 复制 icon.ico
)

echo.
echo ============================================
echo 构建完成！
echo ============================================
echo.
echo 输出目录: %CD%\%BUILD_DIR%
echo 可执行文件: %EXAMPLE_NAME%.exe
echo.
echo 使用说明:
echo 1. 将 %BUILD_DIR% 目录复制到 PIME 的 go 目录
echo 2. 在 backends.json 中添加后端配置
echo 3. 重启 PIME 服务
echo.
pause
