@echo off
setlocal enabledelayedexpansion

REM --- Configuration ---
SET REPO_OWNER=nazar256
SET REPO_NAME=user-prompt-mcp
SET CLIENT_BINARY_NAME=user-prompt-mcp
SET SERVER_BINARY_NAME=user-prompt-server
SET API_REPO_URL=https://api.github.com/repos/%REPO_OWNER%/%REPO_NAME%
SET DOWNLOAD_REPO_URL_BASE=https://github.com/%REPO_OWNER%/%REPO_NAME%

REM --- Default values ---
SET "VERSION_TO_INSTALL=latest"
SET "COMPONENT_TO_INSTALL=all"
SET "OS_NAME=windows"

REM --- Argument Parsing ---
:parse_args
IF "%~1"=="" GOTO :args_done

IF /I "%~1"=="-v" (SET "VERSION_TO_INSTALL=%~2" & SHIFT & SHIFT & GOTO :parse_args)
IF /I "%~1"=="--version" (SET "VERSION_TO_INSTALL=%~2" & SHIFT & SHIFT & GOTO :parse_args)
IF /I "%~1"=="--component" (SET "COMPONENT_TO_INSTALL=%~2" & SHIFT & SHIFT & GOTO :parse_args)
IF /I "%~1"=="-h" (CALL :show_help & exit /b 0)
IF /I "%~1"=="--help" (CALL :show_help & exit /b 0)
IF /I "%~1"=="?" (CALL :show_help & exit /b 0)

echo Unknown option: %~1
CALL :show_help
exit /b 1
:args_done

REM --- Determine Version ---
IF /I "%VERSION_TO_INSTALL%"=="latest" (
    echo Fetching latest release version...
    SET POWERSHELL_COMMAND=(Invoke-RestMethod -Uri '%API_REPO_URL%/releases/latest' -UseBasicParsing).tag_name
    FOR /F "tokens=*" %%i IN ('powershell -NoProfile -ExecutionPolicy Bypass -Command "%POWERSHELL_COMMAND%" 2^>nul') DO SET "VERSION_TO_INSTALL=%%i"
    IF "!VERSION_TO_INSTALL!"=="" (
        echo ERROR: Could not fetch latest release version. Please specify a version with -v.
        echo Make sure PowerShell is working and can access GitHub.
        exit /b 1
    )
    IF /I "!VERSION_TO_INSTALL!"=="latest" (
        REM PowerShell command likely failed silently or returned empty
        echo ERROR: Failed to determine latest version from GitHub.
        exit /b 1
    )
)
echo Installing version: %VERSION_TO_INSTALL%

REM --- Detect Architecture ---
SET ARCH="unknown"
IF "%PROCESSOR_ARCHITECTURE%"=="AMD64" SET ARCH="amd64"
IF "%PROCESSOR_ARCHITECTURE%"=="ARM64" SET ARCH="arm64"

REM Check for 32-bit cmd on 64-bit OS
IF DEFINED PROCESSOR_ARCHITEW6432 (
    IF "%PROCESSOR_ARCHITEW6432%"=="AMD64" SET ARCH="amd64"
    IF "%PROCESSOR_ARCHITEW6432%"=="ARM64" SET ARCH="arm64"
)

IF "!ARCH!"=="unknown" (
    echo ERROR: Unsupported architecture: %PROCESSOR_ARCHITECTURE%. This script supports AMD64 and ARM64.
    exit /b 1
)
SET PLATFORM=%OS_NAME%-!ARCH!
echo Detected platform: !PLATFORM!

REM --- Installation Directory ---
SET INSTALL_DIR=%USERPROFILE%\.local\bin
IF NOT EXIST "%USERPROFILE%\.local" MKDIR "%USERPROFILE%\.local"
IF ERRORLEVEL NEQ 0 (
    echo ERROR: Failed to create directory %USERPROFILE%\.local
    exit /b 1
)
IF NOT EXIST "%INSTALL_DIR%" MKDIR "%INSTALL_DIR%"
IF ERRORLEVEL NEQ 0 (
    echo ERROR: Failed to create directory %INSTALL_DIR%
    exit /b 1
)
echo Selected installation directory: %INSTALL_DIR%

REM --- Determine Binaries to Install ---
SET BINARIES_TO_INSTALL=
IF /I "%COMPONENT_TO_INSTALL%"=="all" (
    SET "BINARIES_TO_INSTALL=%CLIENT_BINARY_NAME% %SERVER_BINARY_NAME%"
) ELSE IF /I "%COMPONENT_TO_INSTALL%"=="client" (
    SET "BINARIES_TO_INSTALL=%CLIENT_BINARY_NAME%"
) ELSE IF /I "%COMPONENT_TO_INSTALL%"=="server" (
    SET "BINARIES_TO_INSTALL=%SERVER_BINARY_NAME%"
) ELSE (
    echo ERROR: Invalid component '%COMPONENT_TO_INSTALL%'. Choose 'client', 'server', or 'all'.
    exit /b 1
)

REM --- Install Binaries ---
SET "INSTALLED_COUNT=0"
FOR %%B IN (%BINARIES_TO_INSTALL%) DO (
    CALL :InstallBinary "%%B" "%VERSION_TO_INSTALL%" "!PLATFORM!" "%INSTALL_DIR%"
    IF ERRORLEVEL NEQ 0 (
        echo ERROR: Failed to install %%B.
        REM No exit here, try to install other components if any
    ) ELSE (
        SET /A INSTALLED_COUNT+=1
    )
)

REM --- Final Messages ---
IF !INSTALLED_COUNT! EQU 0 (
    echo No components were successfully installed.
    exit /b 1
)

echo.
echo Successfully installed:
IF EXIST "%INSTALL_DIR%\%CLIENT_BINARY_NAME%.exe" echo   - %CLIENT_BINARY_NAME%.exe
IF EXIST "%INSTALL_DIR%\%SERVER_BINARY_NAME%.exe" echo   - %SERVER_BINARY_NAME%.exe
echo   (at %INSTALL_DIR%)
echo.
echo IMPORTANT: Ensure "%INSTALL_DIR%" is in your system's PATH.
echo You can add it temporarily for the current command prompt session by running:
echo   set PATH=%%PATH%%;%INSTALL_DIR%
echo.
echo To add it permanently for your user account (requires a new command prompt to take effect):
echo   setx PATH "%%PATH%%;%INSTALL_DIR%"
echo.
echo If curl, tar, or PowerShell commands failed, please ensure they are installed and in your PATH.
exit /b 0


REM --- Subroutine: InstallBinary ---
:InstallBinary <target_binary_name> <version> <platform> <install_dir>
    SET "TARGET_BINARY_NAME=%~1"
    SET "VERSION=%~2"
    SET "PLATFORM=%~3"
    SET "INSTALL_DIR=%~4"
    SET "TARGET_EXE_NAME=%TARGET_BINARY_NAME%.exe"

    echo --- Installing %TARGET_BINARY_NAME% %VERSION% ---

    SET "DOWNLOAD_FILE_BASE_NAME=%TARGET_BINARY_NAME%-%PLATFORM%"
    SET "BINARY_GZ_URL=%DOWNLOAD_REPO_URL_BASE%/releases/download/%VERSION%/%DOWNLOAD_FILE_BASE_NAME%.exe.gz"
    SET "CHECKSUM_URL=%DOWNLOAD_REPO_URL_BASE%/releases/download/%VERSION%/%DOWNLOAD_FILE_BASE_NAME%.exe.sha256"

    SET "TEMP_DIR=%TEMP%\user-prompt-install-%RANDOM%"
    IF EXIST "%TEMP_DIR%" RD /S /Q "%TEMP_DIR%" >nul 2>&1
    MKDIR "%TEMP_DIR%"
    IF ERRORLEVEL NEQ 0 (echo ERROR: Could not create temp directory %TEMP_DIR% & exit /b 1)

    echo Downloading %TARGET_BINARY_NAME% from %BINARY_GZ_URL%...
    curl -L -s --fail -o "%TEMP_DIR%\%DOWNLOAD_FILE_BASE_NAME%.exe.gz" "%BINARY_GZ_URL%"
    IF ERRORLEVEL NEQ 0 (echo ERROR: Failed to download binary. Check URL and network. & RD /S /Q "%TEMP_DIR%" & exit /b 1)

    echo Downloading checksum from %CHECKSUM_URL%...
    curl -L -s --fail -o "%TEMP_DIR%\%DOWNLOAD_FILE_BASE_NAME%.exe.sha256" "%CHECKSUM_URL%"
    IF ERRORLEVEL NEQ 0 (echo ERROR: Failed to download checksum. Check URL and network. & RD /S /Q "%TEMP_DIR%" & exit /b 1)

    echo Decompressing %DOWNLOAD_FILE_BASE_NAME%.exe.gz...
    tar -xzf "%TEMP_DIR%\%DOWNLOAD_FILE_BASE_NAME%.exe.gz" -C "%TEMP_DIR%"
    IF ERRORLEVEL NEQ 0 (echo ERROR: Failed to decompress. Is tar.exe in PATH and functional? & RD /S /Q "%TEMP_DIR%" & exit /b 1)

    SET "DECOMPRESSED_FILE_PATH=%TEMP_DIR%\%DOWNLOAD_FILE_BASE_NAME%.exe"
    IF NOT EXIST "%DECOMPRESSED_FILE_PATH%" (
      echo ERROR: Decompressed file %DECOMPRESSED_FILE_PATH% not found.
      RD /S /Q "%TEMP_DIR%"
      exit /b 1
    )

    echo Verifying checksum for %DECOMPRESSED_FILE_PATH%...
    SET "EXPECTED_HASH="
    FOR /F "tokens=1" %%s IN ('type "%TEMP_DIR%\%DOWNLOAD_FILE_BASE_NAME%.exe.sha256"') DO SET "EXPECTED_HASH=%%s"
    IF "!EXPECTED_HASH!"=="" (echo ERROR: Could not read expected hash from .sha256 file. & RD /S /Q "%TEMP_DIR%" & exit /b 1)

    SET "ACTUAL_HASH="
    FOR /F "skip=1 tokens=*" %%i IN ('certutil -hashfile "%DECOMPRESSED_FILE_PATH%" SHA256') DO (
        FOR /F "tokens=*" %%j IN ("%%i") DO SET "ACTUAL_HASH=%%j" & GOTO :hash_extracted_sub
    )
    :hash_extracted_sub
    SET "ACTUAL_HASH_NO_SPACE=!ACTUAL_HASH: =!"
    IF "!ACTUAL_HASH_NO_SPACE!"=="" (echo ERROR: Could not calculate actual hash using certutil. & RD /S /Q "%TEMP_DIR%" & exit /b 1)

    IF NOT "!EXPECTED_HASH!"=="!ACTUAL_HASH_NO_SPACE!" (
        echo ERROR: Checksum mismatch for %DECOMPRESSED_FILE_PATH%.
        echo   Expected: !EXPECTED_HASH!
        echo   Actual:   !ACTUAL_HASH_NO_SPACE!
        RD /S /Q "%TEMP_DIR%"
        exit /b 1
    )
    echo Checksum verified.

    echo Installing %TARGET_EXE_NAME% to %INSTALL_DIR%...
    MOVE /Y "%DECOMPRESSED_FILE_PATH%" "%INSTALL_DIR%\%TARGET_EXE_NAME%"
    IF ERRORLEVEL NEQ 0 (echo ERROR: Failed to move binary to %INSTALL_DIR%. & RD /S /Q "%TEMP_DIR%" & exit /b 1)

    RD /S /Q "%TEMP_DIR%"
    echo Successfully installed %TARGET_EXE_NAME% to %INSTALL_DIR%.
GOTO :EOF


REM --- Subroutine: Show Help ---
:show_help
    echo Usage: %~nx0 [options]
    echo.
    echo Options:
    echo   -v, --version VERSION    Install specific version (default: latest)
    echo   --component NAME         Specify component to install: 'client', 'server', or 'all' (default: all)
    echo   -h, --help, /?           Show this help message
    echo.
    echo Examples:
    echo   %~nx0
    echo   %~nx0 -v v1.0.0
    echo   %~nx0 --component client -v v0.1.0
    echo.
    echo Prerequisites:
    echo   - curl.exe (for downloads, standard in recent Windows)
    echo   - tar.exe (for decompression, often with Git for Windows or newer Windows)
    echo   - powershell.exe (for fetching latest version, standard in Windows)
    echo   - certutil.exe (for checksums, standard in Windows)
GOTO :EOF
