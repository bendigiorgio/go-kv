
.PHONY: build-local build templ notify-templ-proxy dev 

-include .env

build-local:
	@go build -o ./bin/main cmd/main/main.go

build-tailwind:
	@npx @tailwindcss/cli -i ./internal/web/static/css/style.css -o ./internal/web/static/css/tailwind.css --minify

build:
	@make build-tailwind
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/main cmd/main/main.go

templ:
	@templ generate --watch --proxy=http://localhost:$(APP_PORT) --proxyport=$(TEMPL_PROXY_PORT) --open-browser=false --proxybind="0.0.0.0"

notify-templ-proxy:
	@templ generate --notify-proxy --proxyport=$(TEMPL_PROXY_PORT)

dev:
	@make templ & sleep 1
	@air -c air.toml && px @tailwindcss/cli -i ./internal/web/static/css/style.css -o ./internal/web/static/css/tailwind.css --watch
