# Camcpy

**Camcpy** transforms your Android phone into a webcam using scrcpy.

## Features

- **Wireless ADB Pairing** - Pair your Android device over WiFi without USB
- **Live Camera Streaming** - Stream camera feed via HLS or WebRTC
- **Configurable Settings** - Adjust camera position (front/back), resolution, and FPS
- **Docker Ready** - One-command deployment with docker-compose

```

## Quick Start

### Docker Deployment

```bash
# Build and start all services
docker-compose up --build

# Access the web interface
open http://localhost:8080
```


| WebRTC | 8889 | WebRTC signaling |
