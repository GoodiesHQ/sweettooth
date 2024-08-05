@echo off
docker run --rm -v C:/code/go/sweettooth/:/src -w /src sqlc/sqlc generate