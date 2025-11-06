live/templ:
	templ generate --watch --proxy="http://localhost:8080" --open-browser=false

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
		# --log.time "true" 
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

live:
	$(MAKE) -j4 live/templ live/server live/tailwind live/sync_assets

.PHONY: live live/templ live/server live/tailwind live/sync_assets

