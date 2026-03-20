# Camcpy

**Camcpy** transforms your Android phone into a webcam using scrcpy, with a sleek web interface for pairing, connecting, and streaming.

## Features

- **Wireless ADB Pairing** - Pair your Android device over WiFi without USB
- **Live Camera Streaming** - Stream camera feed via RTSP, HLS, or WebRTC
- **Configurable Settings** - Adjust camera position (front/back), resolution, and FPS
- **Web-Based UI** - Modern interface built with TailwindCSS and daisyUI
- **Docker Ready** - One-command deployment with docker-compose

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Web Browser                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Camcpy Web Server (Go)                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐│
│  │   Pair UI   │  │  Connect UI  │  │  Camera Settings UI     ││
│  │  /pair      │  │  /connect    │  │  /setupcamera            ││
│  └─────────────┘  └─────────────┘  └─────────────────────────┘│
│                              │                                   │
│  ┌───────────────────────────┴───────────────────────────────┐  │
│  │              Handlers (ADB + scrcpy control)              │  │
│  └───────────────────────────┬───────────────────────────────┘  │
└──────────────────────────────┼───────────────────────────────────┘
                               │ ADB Commands
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Android Device                               │
│                     (via ADB WiFi)                               │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼ Camera Stream
┌─────────────────────────────────────────────────────────────────┐
│                      scrcpy (Camera Capture)                     │
│              --video-source=camera --no-window                   │
└──────────────────────────────┬───────────────────────────────────┘
                               │ H.264/AAC Stream
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      MediaMTX Server                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────┐  │
│  │   RTSP   │  │   HLS    │  │ WebRTC   │  │     SRT        │  │
│  │ :8554    │  │  :8888   │  │  :8889   │  │    :8890       │  │
│  └──────────┘  └──────────┘  └──────────┘  └────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Go 1.25+
- Node.js & npm
- [Android SDK Platform Tools](https://developer.android.com/studio/releases/platform-tools) (adb)
- [scrcpy](https://github.com/Genymobile/scrcpy) (with camera support)
- Docker & Docker Compose (for containerized deployment)

### Installing scrcpy with Camera Support

```bash
# Install dependencies
sudo apt install ffmpeg libsdl2-2.0-0 libusb-1.0-0 meson ninja-build gcc git

# Clone and build scrcpy with camera support
git clone -b turn-off-listening https://github.com/rudy-bundela/scrcpy-wip
cd scrcpy-wip
meson x --buildtype=release --strip -Db_lto=true
cd x
ninja
sudo ninja install
```

## Quick Start

### Local Development

```bash
# Install all dependencies
make install

# Run live development server (Go + Templ + Tailwind)
make live
```

Visit `http://localhost:8080`

### Docker Deployment

```bash
# Build and start all services
docker-compose up --build

# Access the web interface
open http://localhost:8080
```

## Usage

### 1. Enable ADB over WiFi on Your Phone

```bash
# Connect via USB first, then:
adb tcpip 5555

# Or use wireless debugging (Android 11+)
```

### 2. Pair Your Device

1. Open Camcpy web interface
2. Navigate to **Pair**
3. Enter your phone's IP address, port (typically `5555`), and pairing code
4. Click "Pair with ADB"

### 3. Connect to Paired Device

1. Navigate to **Connect**
2. Enter IP address and port
3. Click "Connect ADB"

### 4. Configure and Stream

1. Select camera position (front/back)
2. Choose FPS and resolution
3. Click "Start Video Stream"
4. Access stream via MediaMTX endpoints:
   - **RTSP**: `rtsp://localhost:8554/stream`
   - **HLS**: `http://localhost:8888/stream/live.m3u8`
   - **WebRTC**: Use the WebRTC URL from MediaMTX

## Project Structure

```
camcpy/
├── main.go              # HTTP server and route definitions
├── components/          # Templ UI components and scrcpy logic
│   ├── camerahelpers.go # Camera stream handling
│   ├── scrcpyparser.go  # scrcpy output parsing
│   └── *.templ          # HTML template files
├── handlers/            # HTTP request handlers
│   ├── handlepair.go    # ADB pairing logic
│   └── handleconnect.go # ADB connection logic
├── services/            # (Reserved for future services)
├── static/              # Static assets
│   ├── css/            # TailwindCSS styles
│   └── js/             # Datastar frontend JS
├── mediamtx_conf/       # MediaMTX configuration
├── Dockerfile           # Main application container
├── mediamtx.Dockerfile  # MediaMTX server container
└── docker-compose.yml   # Multi-container orchestration
```

## Configuration

### MediaMTX Ports

| Protocol | Port | Purpose |
|----------|------|---------|
| RTSP | 8554 | Standard streaming |
| HLS | 8888 | HTTP Live Streaming |
| WebRTC | 8889 | WebRTC signaling |
| UDP | 8189 | WebRTC ICE candidates |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TEMPL_DEV_MODE` | `1` | Enable development mode (hot reload) |

## Development

### Makefile Commands

```bash
make install       # Install all dependencies (Go, Templ, Tailwind)
make live          # Run live development server with hot reload
make live/templ    # Watch and generate templ components
make live/server   # Watch and rebuild Go server
make live/tailwind # Watch and rebuild CSS
```

### Build for Production

```bash
# Generate static assets
templ generate
npx tailwindcss -i ./static/css/style.css -o ./static/css/tailwind.css --minify

# Build Go binary with embedded assets
CGO_ENABLED=0 go build -v -tags release -o camcpy .
```

## Troubleshooting

### Device Not Found
- Ensure ADB WiFi is enabled: `adb tcpip 5555`
- Check firewall allows port 5555
- Verify device IP: `adb shell ip route`

### Streaming Fails
- Confirm scrcpy supports camera: `scrcpy --list-camera-sizes`
- Check MediaMTX logs: `docker logs mediamtx`
- Verify stream URL matches configuration

### Pairing Timeout
- Ensure pairing code is entered correctly
- Some devices require USB pairing first
- Check device screen for confirmation dialog

## License

MIT
