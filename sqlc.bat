@echo off
REM set "CURRENT_DIR=%~dp0"
REM docker run --rm -v "%CURRENT_DIR%/":/src -w /src sqlc/sqlc generate
docker run --rm -v C:/code/go/sweettooth/:/src -w /src sqlc/sqlc generate