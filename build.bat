@echo off
setlocal

:: Set the name of the output executable
set "EXECUTABLE_NAME=aperture.exe"

echo [APERTURE] Cleaning previous build...
:: Delete the old executable if it exists.
:: The '2>nul' suppresses the "file not found" error if it doesn't exist.
del %EXECUTABLE_NAME% 2>nul

echo [APERTURE] Starting Go build... (using #cgo directives from main.go)
echo.

:: Run the Go build command.
:: -v (verbose) flag shows the packages being compiled, including the cgo commands.
:: -o sets the output file name.
go build -v -o %EXECUTABLE_NAME%

:: Check if the build was successful
if %errorlevel% equ 0 (
    echo.
    echo [APERTURE] BUILD SUCCESSFUL!
    echo [APERTURE] The executable '%EXECUTABLE_NAME%' has been created.
    echo [APERTURE] Run it by typing: %EXECUTABLE_NAME%
) else (
    echo.
    echo [APERTURE] BUILD FAILED.
    echo [APERTURE] Please review the compilation or linker errors above.
)

echo.
endlocal
pause