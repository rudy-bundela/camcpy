# Deployment Guide

This guide covers deploying Camcpy to production using Docker.

## Docker Architecture

Camcpy uses a multi-container setup:

```yaml
services:
  web:           # Go web server + scrcpy
    image: camcpy:latest
    ports: 8080
    
  mediamtx:      # RTSP/HLS/WebRTC server
    image: mediamtx:latest
    ports: 8554, 8888, 8889, 8189/udp, 8890/udp

networks:
  scrcpy_net:    # Bridge network for container communication
  
volumes:
  scrcpy_adb_keys:  # Persistent ADB pairing keys
```

## Quick Start (Docker Compose)

### Build and Run

```bash
# Build images
docker-compose build

# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Access Services

| Service | URL |
|---------|-----|
| Web UI | http://localhost:8080 |
| RTSP | rtsp://localhost:8554/stream |
| HLS | http://localhost:8888/stream/live.m3u8 |
| WebRTC | http://localhost:8889/stream |
| SRT | srt://localhost:8890/stream |

## Dockerfile Breakdown

### Main Dockerfile (camcpy)

Three-stage build:

**Stage 1: Build scrcpy**
```dockerfile
FROM alpine AS scrcpy_builder
# Installs: ffmpeg, gcc, meson, libusb, android-tools, sdl2
# Clones custom scrcpy fork with camera support
# Builds with: meson + ninja
```

**Stage 2: Build Go server**
```dockerfile
FROM golang:1.25-alpine AS webserver_builder
# Installs: templ, npm
# Generates: HTML from .templ files
# Builds: CSS from TailwindCSS
# Compiles: Go binary with -tags release
```

**Stage 3: Runtime**
```dockerfile
FROM alpine:latest
# Copies: scrcpy binaries from stage 1
# Copies: Go binary from stage 2
# Installs: ffmpeg, libusb, sdl2, android-tools
```

### MediaMTX Dockerfile

Lightweight Alpine container:

```dockerfile
FROM alpine
# Downloads: MediaMTX v1.14.0 binary
# Entrypoint: ./mediamtx with config mount
```

## Configuration

### MediaMTX Configuration

The `mediamtx_conf/mediamtx.yml` file is mounted into the container:

```bash
volumes:
  - ./mediamtx_conf:/mediamtx_conf
```

#### Key Settings

**RTSP Server:**
```yaml
rtsp: yes
rtspAddress: :8554
rtspTransports: [udp, multicast, tcp]
```

**HLS Streaming:**
```yaml
hls: yes
hlsAddress: :8888
hlsVariant: lowLatency
hlsSegmentCount: 7
hlsSegmentDuration: 1s
```

**WebRTC:**
```yaml
webrtc: yes
webrtcAddress: :8889
webrtcLocalUDPAddress: :8189
webrtcAdditionalHosts: [192.168.0.241]  # CHANGE THIS
```

### Customizing MediaMTX

Edit `mediamtx_conf/mediamtx.yml`:

1. **Change WebRTC external IP:**
   ```yaml
   webrtcAdditionalHosts: [YOUR_SERVER_IP]
   ```

2. **Add authentication:**
   ```yaml
   authMethod: internal
   authInternalUsers:
     - user: streamuser
       pass: streampass
       permissions:
         - action: read
         - action: playback
   ```

3. **Enable recording:**
   ```yaml
   pathDefaults:
     record: yes
     recordPath: ./recordings/%path/%Y-%m-%d_%H-%M-%S
   ```

## Network Configuration

### Ports

| Port | Protocol | Service | Purpose |
|------|----------|---------|---------|
| 8080 | TCP | web | Web UI |
| 8554 | TCP | mediamtx | RTSP |
| 8888 | TCP | mediamtx | HLS |
| 8889 | TCP | mediamtx | WebRTC HTTP |
| 8189 | UDP | mediamtx | WebRTC ICE |
| 8890 | UDP | mediamtx | SRT |

### Firewall Rules

```bash
# Ubuntu/Debian with ufw
ufw allow 8080/tcp   # Web UI
ufw allow 8554/tcp   # RTSP
ufw allow 8888/tcp   # HLS
ufw allow 8889/tcp   # WebRTC
ufw allow 8189/udp   # WebRTC ICE
ufw allow 8890/udp   # SRT
```

## Production Build (Manual)

### Without Docker

```bash
# Generate templ templates
templ generate

# Build TailwindCSS
npx tailwindcss -i ./static/css/style.css -o ./static/css/tailwind.css --minify

# Build Go binary
CGO_ENABLED=0 go build -v -tags release -o camcpy .

# Run
./camcpy
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TEMPL_DEV_MODE` | unset | Set to `1` for development mode |

## Persistence

### ADB Pairing Keys

Pairing keys are stored in Docker volume:

```yaml
volumes:
  - scrcpy_adb_keys:/root/.android
```

Devices paired once remain paired across restarts.

### View Volume Contents

```bash
# Inspect volume
docker volume inspect camcpy_scrcpy_adb_keys

# Copy contents
docker run --rm -v camcpy_scrcpy_adb_keys:/data -v $(pwd):/backup alpine tar czf /backup/adb-keys.tar.gz -C /data .
```

## Resource Usage

### Memory

- **web container**: ~100-200MB (scrcpy overhead)
- **mediamtx container**: ~50MB baseline, scales with connections

### CPU

- **Idle**: Minimal
- **Streaming**: Depends on resolution/FPS, 10-30% typical

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker-compose logs web
docker-compose logs mediamtx

# Common issues:
# - Port already in use
# - Volume permissions
# - Missing dependencies
```

### WebRTC Not Working

1. **Check webrtcAdditionalHosts** in mediamtx.yml matches server IP
2. **Verify UDP port 8189** is open
3. **Test with STUN**: Uses `stun:stun.l.google.com:19302`

### Streaming Quality Issues

1. **Reduce FPS/resolution** in camera settings
2. **Check network** between server and phone
3. **Adjust MediaMTX buffers:**
   ```yaml
   writeQueueSize: 8192  # Higher = better throughput
   udpMaxPayloadSize: 1472  # Adjust for network MTU
   ```

### Device Disconnects

1. **Keep screen on**: scrcpy needs screen to capture
2. **Check WiFi signal**: Weak connection = drops
3. **ADB timeout**: Increase if on unreliable network

## Reverse Proxy

### Nginx

```nginx
server {
    listen 443 ssl;
    server_name camcpy.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    location /stream/ {
        proxy_pass http://localhost:8888/;
        proxy_http_version 1.1;
    }
}
```

### Caddy

```caddy
camcpy.example.com {
    reverse_proxy localhost:8080
}

stream.example.com {
    reverse_proxy localhost:8888
}
```

## Security

### Production Checklist

- [ ] Change default ports if needed
- [ ] Enable MediaMTX authentication
- [ ] Use HTTPS for web UI
- [ ] Firewall off unused ports
- [ ] Regular container updates
- [ ] Monitor resource usage

### Hardening

1. **Run as non-root:**
   ```dockerfile
   RUN adduser -D appuser
   USER appuser
   ```

2. **Read-only rootfs:**
   ```yaml
   web:
     read_only: true
     volumes:
       - /tmp:/tmp  # scrcpy needs temp space
   ```

3. **Limit resources:**
   ```yaml
   deploy:
     resources:
       limits:
         memory: 512M
   ```

## Updates

### Update Camcpy

```bash
git pull
docker-compose build
docker-compose up -d
```

### Update MediaMTX

Edit `mediamtx.Dockerfile`:
```dockerfile
RUN curl -L -o /mediamtx_v1.X.X_linux_amd64.tar.gz \
    https://github.com/bluenviron/mediamtx/releases/download/v1.X.X/mediamtx_v1.X.X_linux_amd64.tar.gz
```

Rebuild:
```bash
docker-compose build mediamtx
docker-compose up -d
```
