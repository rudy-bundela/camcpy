# Architecture Guide

This document describes Camcpy's system architecture for developers contributing to the project.

## High-Level Overview

Camcpy is a Go web server that bridges Android's ADB/WiFi debugging protocol with scrcpy camera capture, streaming the output through MediaMTX.

```
┌─────────────┐     HTTP/SSE      ┌─────────────────┐
│   Browser   │ ←───────────────→ │   Go Server     │
│  (Frontend) │                   │   (Camcpy)      │
└─────────────┘                   └────────┬────────┘
                                           │
                    ┌──────────────────────┼──────────────────────┐
                    │                      │                      │
                    ▼                      ▼                      ▼
            ┌───────────────┐      ┌───────────────┐      ┌───────────────┐
            │ adb (pairing) │      │ adb (connect) │      │    scrcpy     │
            │  Port 5555    │      │   Port 5555   │      │ (camera grab) │
            └───────────────┘      └───────────────┘      └───────┬───────┘
                    │                      │                      │
                    └──────────────────────┼──────────────────────┘
                                           │
                    ┌──────────────────────┴──────────────────────┐
                    │          Android Device (WiFi)             │
                    │    [Pairing] → [Connection] → [Streaming]  │
                    └─────────────────────────────────────────────┘
                                           │
                                           ▼ Camera Stream
                                  ┌─────────────────┐
                                  │    MediaMTX     │
                                  │ RTSP/HLS/WebRTC │
                                  └─────────────────┘
```

## Core Components

### 1. Web Server (`main.go`)

The Go HTTP server handles all routing and middleware.

**Key Responsibilities:**
- Serve static assets (CSS, JS)
- Route HTTP requests to handlers
- Initialize ScrcpyInfo singleton
- Middleware for cache control in dev mode

### 2. Components Package (`components/`)

#### ScrcpyInfo Struct
The central state holder for camera configuration:

```go
type ScrcpyInfo struct {
    DeviceName     string             // Android device model
    Cameras        []Camera           // Available cameras
    cancelStream   context.CancelFunc // Stream termination
    ActiveSettings DatastarSignalsStruct // Current stream config
}
```

#### Camera Struct
Represents a camera on the device:

```go
type Camera struct {
    ID       string       // e.g., "0", "1"
    Position string       // "front" or "back"
    Sizes    []SizeConfig // Supported resolutions
}
```

#### SizeConfig Struct
Resolution and FPS capabilities:

```go
type SizeConfig struct {
    Resolution string      // "1920x1080"
    Width      int         // 1920
    Height     int         // 1080
    FPS        []FPSOption // Available frame rates
}
```

### 3. Handlers Package (`handlers/`)

#### HandlePairing (`handlepair.go`)
Handles ADB device pairing over WiFi.

**Flow:**
1. Receive IP, port, and pairing code from form
2. Execute `adb pair <ip:port> <code>`
3. Display output in CodePen component
4. Redirect to connect page on success

#### HandleADBConnect (`handleconnect.go`)
Connects to a previously paired device.

**Flow:**
1. Receive IP and port from form
2. Execute `adb connect <ip:port>`
3. Display connection status
4. Redirect to camera setup on success

### 4. Frontend Components (`.templ` files)

Camcpy uses [templ](https://templ.guide/) for type-safe Go-based HTML templates.

**Key Components:**

| Component | File | Purpose |
|-----------|------|---------|
| Layout | `layout.templ` | Base HTML structure, CSS/JS includes |
| Navbar | `navbar.templ` | Navigation between pages |
| PairForm | `pairForm.templ` | Device pairing form |
| ConnectForm | `connectForm.templ` | Device connection form |
| CameraOptions | `cameraoptions.templ` | Camera selection UI |

### 5. Datastar Integration

Camcpy uses [Datastar](https://github.com/starfederation/datastar) for reactive UI without a JavaScript framework.

**How It Works:**
1. Initial HTML rendered by templ
2. Datastar JS loaded from `/static/js/datastar.js`
3. User interactions trigger SSE requests to Go handlers
4. Go handlers return Datastar SSE commands
5. Datastar patches the DOM with new content

**Signal Structure:**
```go
type DatastarSignalsStruct struct {
    Position   string // "front" or "back"
    Fps        int    // Target FPS
    Resolution string // Resolution label
    CamID      string // Camera identifier
}
```

## Data Flow: Streaming a Camera

### Step 1: Fetch Available Cameras
```
Browser → GET /setupcamerasse → RunGetScrcpyDetails()
                                    │
                                    ▼ scrcpy --list-camera-sizes
                              ParseScrcpyOutput() → ScrcpyInfo struct
                                    │
                                    ▼ SSE response with camera options
Browser displays camera selection UI
```

### Step 2: Start Streaming
```
User clicks "Start Stream"
Browser → POST /camera/startstream → HandleStartStream()
                                        │
                                        ▼ Build scrcpy command
                              scrcpy --video-source=camera \
                                    --camera-id=0 \
                                    --camera-size=1920x1080 \
                                    --camera-fps=30 \
                                    -ra.mp4
                                        │
                                        ▼ Camera stream → MediaMTX
Browser shows "Streaming..." status
```

### Step 3: Stop Streaming
```
User clicks "Stop Stream"
Browser → POST /camera/stopstream → HandleStopStream()
                                        │
                                        ▼ cancel context
                              Stream process terminated
```

## Development vs Production Builds

Camcpy uses Go build tags to switch between dev and prod configurations.

### Dev Mode (`//go:build !release`)
```go
const IsDev = true

func GetBaseScrcpyArgs() []string {
    return []string{
        "--video-source=camera",
    }
}
```

### Production Mode (`//go:build release`)
```go
const IsDev = false

func GetBaseScrcpyArgs() []string {
    return []string{
        "-ra.mp4",           // Output to file
        "--no-playback",     // No GUI window
        "--video-source=camera",
        "--no-window",       // No window
        "--no-control",      // No control
        "--audio-codec=aac", // AAC audio
    }
}
```

## MediaMTX Streaming

MediaMTX receives the camera stream and provides multiple output formats:

| Protocol | URL | Latency | Use Case |
|----------|-----|---------|----------|
| RTSP | `rtsp://localhost:8554/stream` | ~500ms | Standard players |
| HLS | `http://localhost:8888/stream/live.m3u8` | ~3-5s | Web playback |
| WebRTC | `http://localhost:8889/stream` | ~200ms | Low-latency web |
| SRT | `srt://localhost:8890/stream` | Variable | Custom apps |

## State Management

Camcpy uses in-memory state via the `ScrcpyInfo` singleton:

```go
scrcpyStruct := components.ScrcpyInfo{}
```

**Note:** This means:
- State is not persisted across restarts
- Only one device can be configured at a time
- Consider adding Redis/database for multi-device support

## Security Considerations

1. **ADB Pairing** - Pairing codes should be kept private
2. **Network Access** - Devices on same network only
3. **No Authentication** - MediaMTX is open (configurable)
4. **Volume Mount** - `scrcpy_adb_keys` Docker volume stores ADB keys

## Future Architecture Improvements

1. **Multi-Device Support** - Per-device state management
2. **WebRTC Frontend** - Native web player integration
3. **Recording** - Save streams to disk
4. **Authentication** - Secure MediaMTX access
5. **Config Persistence** - Save preferences to database
