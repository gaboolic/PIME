@echo off
setlocal

echo ============================================
echo  PIME Go Backend Build Script
echo ============================================
echo.

set "ROOT_DIR=%~dp0"
if "%ROOT_DIR:~-1%"=="\" set "ROOT_DIR=%ROOT_DIR:~0,-1%"
set "BUILD_ROOT=%ROOT_DIR%\build"
set "PACKAGE_DIR=%BUILD_ROOT%\go-backend"
set "SERVER_EXE=%PACKAGE_DIR%\server.exe"
set "BACKEND_SNIPPET=%BUILD_ROOT%\backends.go-backend.json"

REM Check Go environment
where go >nul 2>nul
if errorlevel 1 (
    echo [ERROR] Go was not found in PATH.
    echo Install Go from: https://golang.org/dl/
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do (
    echo [INFO] Go version: %%i
)

echo.
echo ============================================
echo Step 1: Prepare output directory
echo ============================================
echo.

if exist "%PACKAGE_DIR%" (
    echo [INFO] Removing old build output: "%PACKAGE_DIR%"
    rmdir /s /q "%PACKAGE_DIR%"
)

mkdir "%PACKAGE_DIR%"
if errorlevel 1 (
    echo [ERROR] Failed to create output directory: "%PACKAGE_DIR%"
    exit /b 1
)

echo [INFO] Output directory: "%PACKAGE_DIR%"

echo.
echo ============================================
echo Step 2: Sync Go dependencies
echo ============================================
echo.

pushd "%ROOT_DIR%"
go mod tidy
if errorlevel 1 (
    echo [WARN] go mod tidy failed, continuing...
)

echo.
echo ============================================
echo Step 3: Build go-backend server
echo ============================================
echo.

set "GOOS=windows"
set "GOARCH=amd64"

echo [INFO] Building server.exe ...
go build -ldflags "-s -w" -o "%SERVER_EXE%" .
if errorlevel 1 (
    echo [ERROR] Failed to build server.exe
    popd
    exit /b 1
)

echo [INFO] Built: "%SERVER_EXE%"

echo.
echo ============================================
echo Step 4: Copy input_methods
echo ============================================
echo.

if not exist "%ROOT_DIR%\input_methods" (
    echo [ERROR] Missing input_methods directory: "%ROOT_DIR%\input_methods"
    popd
    exit /b 1
)

xcopy "%ROOT_DIR%\input_methods" "%PACKAGE_DIR%\input_methods\" /E /I /Y >nul
if errorlevel 1 (
    echo [ERROR] Failed to copy input_methods
    popd
    exit /b 1
)

echo [INFO] input_methods copied

echo.
echo ============================================
echo Step 5: Generate backends.json snippet
echo ============================================
echo.

> "%BACKEND_SNIPPET%" echo [
>> "%BACKEND_SNIPPET%" echo   {
>> "%BACKEND_SNIPPET%" echo     "name": "go-backend",
>> "%BACKEND_SNIPPET%" echo     "command": "go-backend\\server.exe",
>> "%BACKEND_SNIPPET%" echo     "workingDir": "go-backend",
>> "%BACKEND_SNIPPET%" echo     "params": ""
>> "%BACKEND_SNIPPET%" echo   }
>> "%BACKEND_SNIPPET%" echo ]

echo [INFO] Generated: "%BACKEND_SNIPPET%"
popd

echo.
echo ============================================
echo Build completed
echo ============================================
echo.
echo Output directory:
echo   "%PACKAGE_DIR%"
echo.
echo Install target:
echo   C:\Program Files (x86)\PIME\go-backend
echo.
echo Notes:
echo 1. backends.json in this repo uses a top-level array.
echo 2. Ensure C:\Program Files (x86)\PIME\backends.json includes go-backend.
echo 3. Ensure C:\Program Files (x86)\PIME\go-backend\input_methods\*\ime.json exists.
echo 4. Re-register both PIMETextService.dll files after copying.
echo 5. Start or restart PIMELauncher.exe after install.
echo.
