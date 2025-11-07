.PHONY: help install/templ install/godeps install/tailwind install live live/templ live/server live/tailwind live/sync_assets

help:  ## Show this help.
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_-]+:.*## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install/templ:
	go install github.com/a-h/templ/cmd/templ@latest
	
install/godeps:
	go get camcpy/components && go mod tidy 

install/tailwind:
	npm install tailwindcss @tailwindcss/cli

install: ## Install dependencies (templ and godeps)
	$(MAKE) install/templ install/godeps install/tailwind

live/templ:
	templ generate --watch --proxy="http://localhost:8080" --proxybind="0.0.0.0" --open-browser=false

live/server:
	TEMPL_DEV_MODE=1 go run github.com/air-verse/air@latest \
		--build.cmd "go build -o tmp/bin/main" --build.bin "tmp/bin/main" --build.delay "100" \
		--build.include_ext "go" \
		--build.exclude_dir "node_modules" \
		--log.time "true" \
		--build.stop_on_error "false" \
		--misc.clean_on_exit true

live/tailwind:
	go run github.com/air-verse/air@latest \
		--build.cmd "npx tailwindcss -i ./static/css/style.css -o ./static/css/tailwind.css --minify" \
		--build.delay "10" \
		--build.exclude_dir "node_modules" \
		--log.time "true" 
		--build.include_ext "templ,go" \
		--color.main "magenta" \
		--color.watcher "cyan" \
		--color.build "yellow" \
		--color.runner "green" \
		--misc.clean_on_exit true

live/sync_assets:
	go run github.com/air-verse/air@latest \
		--build.cmd "templ generate --notify-proxy" \
		--build.bin "true" \
		--build.exclude_dir "node_modules" \
		--build.delay "100" \
		--log.time "true" \
		--build.include_ext "css" \
		--misc.clean_on_exit true

live: ## Run the dev server
	$(MAKE) -j4 live/templ live/server live/tailwind live/sync_assets


