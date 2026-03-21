# Development Guide

This guide covers everything you need to start developing Camcpy.

## Prerequisites

### Required Tools

1. **Go 1.25+** - [Install](https://go.dev/doc/install)
2. **Node.js & npm** - [Install](https://nodejs.org/)
3. **adb** (Android SDK Platform Tools) - [Install](https://developer.android.com/studio/releases/platform-tools)
4. **scrcpy** with camera support - See [scrcpy setup](#scrcpy-setup)

### Optional Tools

- **air** - Live reload for Go (`go install github.com/air-verse/air@latest`)
- **templ** - Type-safe templates (`go install github.com/a-h/templ/cmd/templ@latest`)

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/camcpy.git
cd camcpy
```

### 2. Install Go Dependencies

```bash
go mod download
```

### 3. Install Node Dependencies

```bash
npm install
```

### 4. Install Development Tools

```bash
make install
```

This runs:
- `make install/templ` - Installs templ CLI
- `make install/godeps` - Installs Go dependencies
- `make install/tailwind` - Installs TailwindCSS and daisyUI

### 5. scrcpy Setup

Camcpy requires scrcpy with camera support. The project uses a [custom fork](https://github.com/rudy-bundela/scrcpy-wip):

```bash
# Install build dependencies (Ubuntu/Debian)
sudo apt install ffmpeg libsdl2-2.0-0 libusb-1.0-0 meson ninja-build gcc git pkg-config

# Clone the fork
git clone -b turn-off-listening https://github.com/rudy-bundela/scrcpy-wip
cd scrcpy-wip

# Build
meson x --buildtype=release --strip -Db_lto=true
cd x
ninja

# Install
sudo ninja install
```

## Running the Development Server

### Using Make

```bash
make live
```

This starts 4 concurrent processes:
- `live/templ` - Watches `.templ` files, regenerates Go code
- `live/server` - Watches `.go` files, hot-reloads the server
- `live/tailwind` - Watches CSS, rebuilds styles
- `live/sync_assets` - Syncs generated assets

### Manual Setup

**Terminal 1 - Generate templ:**
```bash
templ generate --watch --proxy="http://localhost:8080" --proxybind="0.0.0.0" --open-browser=false
```

**Terminal 2 - TailwindCSS:**
```bash
npx tailwindcss -i ./static/css/style.css -o ./static/css/tailwind.css --minify --watch
```

**Terminal 3 - Go server:**
```bash
TEMPL_DEV_MODE=1 air
```

## Project Structure

```
camcpy/
├── main.go                    # Entry point, routes
├── components/
│   ├── *.templ               # HTML templates
│   ├── *_templ.go            # Generated Go code (don't edit)
│   ├── *helpers.go           # Template helper functions
│   ├── scrcpyparser.go       # scrcpy output parsing
│   ├── config_dev.go         # Dev configuration
│   └── config_prod.go        # Production configuration
├── handlers/
│   ├── handlepair.go         # ADB pairing handler
│   └── handleconnect.go      # ADB connect handler
├── static/
│   ├── css/
│   │   └── style.css         # TailwindCSS source
│   └── js/
│       └── datastar.js       # Frontend reactivity
├── mediamtx_conf/
│   └── mediamtx.yml          # MediaMTX configuration
├── Dockerfile                # Multi-stage build
├── mediamtx.Dockerfile       # MediaMTX container
└── docker-compose.yml        # Orchestration
```

## Coding Conventions

### Go Code

1. **Error Handling** - Always log errors, don't panic
   ```go
   if err != nil {
       log.Println("Error doing thing: ", err)
       return
   }
   ```

2. **Context Usage** - Use context for cancellation
   ```go
   ctx, cancel := context.WithCancel(context.Background())
   defer cancel()
   ```

3. **SSE Responses** - Use Datastar helpers
   ```go
   sse := datastar.NewSSE(w, r)
   if err := sse.PatchElementTempl(component); err != nil {
       log.Println(err)
   }
   ```

### Templ Templates

1. **Naming** - PascalCase component names
   ```templ
   templ MyComponent() { ... }
   ```

2. **Props** - Pass as function parameters
   ```templ
   templ UserCard(name string, age int) { ... }
   ```

3. **Conditional** - Use `if` blocks
   ```templ
   if showButton {
       <button>Click</button>
   }
   ```

4. **Loops** - Use `for` with range
   ```templ
   for _, item := range items {
       <div>{ item }</div>
   }
   ```

### Datastar Frontend

1. **Fetching** - Use `data-on:click` or `data-on:change`
   ```html
   <button data-on:click="@get('/endpoint')">Submit</button>
   ```

2. **Form Submission** - Set contentType
   ```html
   <button data-on:click="@post('/endpoint', { contentType: 'form'})">Submit</button>
   ```

3. **Signal Binding** - Use `data-bind`
   ```html
   <input data-bind:fps />
   ```

4. **Conditional Display** - Use `data-show`
   ```html
   <div data-show="$fps != '0'">Ready</div>
   ```

## Build for Production

### Generate Assets

```bash
templ generate
npx tailwindcss -i ./static/css/style.css -o ./static/css/tailwind.css --minify
```

### Build Binary

```bash
CGO_ENABLED=0 go build -v -tags release -o camcpy .
```

### Test Docker Build

```bash
docker build -t camcpy:latest .
docker-compose up --build
```

## Testing

### Manual Testing

1. **Pair Flow**
   - Start server: `make live`
   - Navigate to `/pair`
   - Enter invalid IP → Check error handling
   - Enter valid IP + code → Verify pairing

2. **Connect Flow**
   - After pairing, go to `/connect`
   - Verify connection to paired device

3. **Camera Stream**
   - Go to `/setupcamera`
   - Verify camera options load
   - Start/stop stream
   - Check MediaMTX endpoints

### Debug Endpoints

| Endpoint | Purpose |
|----------|---------|
| `/printstruct` | Debug ScrcpyInfo state |
| `/setupcamerasse` | View camera options |

## Common Issues

### templ generate fails

```bash
# Check templ is installed
templ version

# Reinstall if needed
go install github.com/a-h/templ/cmd/templ@latest
```

### TailwindCSS not updating

```bash
# Clear cache
rm -rf node_modules/.cache

# Rebuild
npx tailwindcss -i ./static/css/style.css -o ./static/css/tailwind.css
```

### adb not found

```bash
# Add to PATH or install android-tools
sudo apt install android-tools-adb
```

### scrcpy camera not working

```bash
# Verify scrcpy has camera support
scrcpy --list-camera-sizes

# Should output camera info, not error
```

## Git Workflow

1. **Branch naming**: `feature/your-feature` or `fix/your-fix`
2. **Commits**: Clear, descriptive messages
3. **PRs**: Reference issues, describe changes

## Code Style

- Run `go fmt` before committing
- Keep functions small and focused
- Add comments for complex logic
- Write tests for new handlers

## Next Steps

- Read [API Documentation](api.md) for endpoint details
- Read [Deployment Guide](deployment.md) for production setup
- Check existing code patterns before adding new features
