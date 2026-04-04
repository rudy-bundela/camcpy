# Camcpy

**Camcpy** transforms your Android phone into a webcam using scrcpy.

## Features

- **Wireless ADB Pairing** - Pair your Android device over WiFi without USB
- **Live Camera Streaming** - Stream camera feed via RTSP, HLS, or WebRTC
- **Configurable Settings** - Adjust camera position (front/back), resolution, and FPS
- **Docker Ready** - One-command deployment with docker-compose

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Web Browser                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Camcpy Web Server (Go)                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   Pair UI   │  │  Connect UI │  │  Camera Settings UI     │  │
│  │  /pair      │  │  /connect   │  │  /setupcamera           │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
│                              │                                  │
│  ┌───────────────────────────┴───────────────────────────────┐  │
│  │              Handlers (ADB + scrcpy control)              │  │
│  └───────────────────────────┬───────────────────────────────┘  │
└──────────────────────────────┼──────────────────────────────────┘
                               │ ADB Commands
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Android Device                              │
│                     (via ADB WiFi)                              │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼ Camera Stream
┌─────────────────────────────────────────────────────────────────┐
│                      scrcpy (Camera Capture)                    │
│              --video-source=camera --no-window                  │
└──────────────────────────────┬──────────────────────────────────┘
                               │ H.264/AAC Stream
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      MediaMTX Server                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────┐   │
│  │   RTSP   │  │   HLS    │  │ WebRTC   │  │     SRT        │   │
│  │ :8554    │  │  :8888   │  │  :8889   │  │    :8890       │   │
│  └──────────┘  └──────────┘  └──────────┘  └────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

### Docker Deployment

```bash
# Build and start all services
docker-compose up --build

# Access the web interface
open http://localhost:8080
```

## Usage

### 1. Enable ADB over WiFi on Your Phone

TODO: Fix this incorrect stuff

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
   - **HLS**: `http://localhost:8888/live/stream`
   - **WebRTC**: `http://localhost:8889/live/stream`

### MediaMTX Ports

| Protocol | Port | Purpose |
|----------|------|---------|
| HLS | 8888 | HTTP Live Streaming |
| WebRTC | 8889 | WebRTC signaling |
