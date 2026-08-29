@echo off
setlocal
cd /d "%~dp0"

where go >nul 2>&1
if errorlevel 1 (
  echo ERROR: go is not on PATH. Install Go 1.25+ from https://go.dev/dl/
  exit /b 1
)

echo Running tests...
go test ./...
if errorlevel 1 (
  echo ERROR: tests failed; refusing to build.
  exit /b 1
)

if not exist dist mkdir dist

echo Building dist\permitdenied.exe...
go build -buildvcs=false -o dist\permitdenied.exe .\cmd\permitdenied
if errorlevel 1 (
  echo ERROR: build failed.
  exit /b 1
)

echo.
echo Built: %cd%\dist\permitdenied.exe
echo Run it to play. Do not commit this exe — CI produces the package.
exit /b 0
