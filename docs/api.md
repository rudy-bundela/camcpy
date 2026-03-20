# API Reference

This document describes all HTTP endpoints, handlers, and data structures in Camcpy.

## HTTP Endpoints

### Page Routes

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/` | `Welcome()` | Landing page |
| GET | `/pair` | `PairformComponent()` | Device pairing form |
| GET | `/connect` | `ConnectformComponent()` | Device connection form |
| GET | `/setupcamera` | `SetupCamera()` | Camera configuration page |

### API Routes

#### Pairing

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/cameraoptions` | `HandlePairing()` | Get camera info (legacy) |
| POST | `/pairingendpoint` | `HandlePairing()` | Pair device via ADB |
| POST | `/camera/idupdate` | `HandleCameraIDUpdate()` | Update selected camera ID |

**POST `/pairingendpoint`**

Form fields:
- `ipaddr` - Device IP address
- `port` - ADB port (default: 5555)
- `code` - Pairing code from device

Response: SSE with CodePen content + redirect to `/connect`

#### Connection

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/mediamtxoptions` | `HandlePairing()` | Legacy endpoint |
| GET | `/monitor` | `HandlePairing()` | Legacy endpoint |
| POST | `/adbconnect` | `HandleADBConnect()` | Connect to paired device |

**POST `/adbconnect`**

Form fields:
- `ipaddr` - Device IP address
- `port` - ADB port (default: 5555)

Response: SSE with connection status + redirect to `/setupcamera`

#### Camera Configuration

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/setupcamerasse` | `HandleGetCameraOptions()` | Fetch camera options |
| POST | `/camera/fpsupdate` | `HandleCameraFPSUpdate()` | Update FPS selection |
| POST | `/camera/resolutionupdate` | `HandleCameraResolutionUpdate()` | Update resolution |
| POST | `/camera/startstream` | `HandleStartStream()` | Start camera stream |
| POST | `/camera/stopstream` | `HandleStopStream()` | Stop camera stream |

**GET `/setupcamerasse`**

Response: SSE patch with:
- Camera options component
- Current signals (position, fps, resolution, camid)

**POST `/camera/fpsupdate`**

Request body: Form with signals
Response: SSE with updated FPS options

**POST `/camera/resolutionupdate`**

Request body: Form with signals
Response: SSE with updated resolution options

**POST `/camera/startstream`**

Request body: Form with signals
Response: SSE status updates, starts scrcpy process

**POST `/camera/stopstream`**

Response: SSE status update, cancels scrcpy process

### Debug Routes

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/printstruct` | `PrintStruct()` | Output ScrcpyInfo as JSON |

## Data Structures

### ScrcpyInfo

Main state holder for camera configuration.

```go
type ScrcpyInfo struct {
    DeviceName     string             `json:"device_name"`
    Cameras        []Camera           `json:"cameras"`
    cancelStream   context.CancelFunc `json:"-"`  // Not serialized
    ActiveSettings DatastarSignalsStruct
}
```

### Camera

Represents a physical camera on the device.

```go
type Camera struct {
    ID       string       `json:"id"`        // e.g., "0", "1"
    Position string       `json:"position"`  // "front" or "back"
    Sizes    []SizeConfig `json:"sizes"`     // Supported resolutions
}
```

### SizeConfig

Resolution and FPS support for a camera.

```go
type SizeConfig struct {
    Resolution string      `json:"resolution"`  // "1920x1080"
    Width      int         `json:"width"`       // 1920
    Height     int         `json:"height"`      // 1080
    FPS        []FPSOption `json:"fps"`         // Supported frame rates
}
```

### FPSOption

Individual FPS option.

```go
type FPSOption struct {
    Value     int  `json:"value"`      // e.g., 30, 60, 120
    HighSpeed bool `json:"high_speed"` // High-speed capture mode
}
```

### ResolutionOption

Resolution option for UI display.

```go
type ResolutionOption struct {
    Value     string  // "1920x1080"
    Label     string  // "1920x1080 (high-speed)"
    HighSpeed bool    // High-speed flag
}
```

### DatastarSignalsStruct

Frontend signal state.

```go
type DatastarSignalsStruct struct {
    Position   string `json:"position"`   // "front" or "back"
    Fps        int    `json:"fps,string"`  // String for Datastar
    Resolution string `json:"resolution"`  // Resolution label
    CamID      string `json:"camid"`       // Camera ID
}
```

## Handler Details

### HandlePairing

Processes ADB pairing requests.

```go
func HandlePairing(w http.ResponseWriter, r *http.Request) {
    // 1. Parse form: ipaddr, port, code
    // 2. Execute: adb pair <ip>:<port> <code>
    // 3. Patch CodePen with output
    // 4. Redirect to /connect on success
}
```

### HandleADBConnect

Connects to a paired device.

```go
func HandleADBConnect(w http.ResponseWriter, r *http.Request) {
    // 1. Parse form: ipaddr, port
    // 2. Execute: adb connect <ip>:<port>
    // 3. Check for "connected to" in output
    // 4. Redirect to /setupcamera on success
}
```

### HandleGetCameraOptions

Fetches and parses camera capabilities.

```go
func (s *ScrcpyInfo) HandleGetCameraOptions(w http.ResponseWriter, r *http.Request) {
    // 1. If DeviceName empty, run scrcpy --list-camera-sizes
    // 2. Parse output into Cameras array
    // 3. Return CameraComponent + signals via SSE
}
```

### HandleStartStream

Starts the scrcpy camera stream.

```go
func (s *ScrcpyInfo) HandleStartStream(w http.ResponseWriter, r *http.Request) {
    // 1. Read signals from request
    // 2. Cancel any existing stream
    // 3. Build scrcpy command with options:
    //    --video-source=camera
    //    --camera-id=<id>
    //    --camera-size=<resolution>
    //    --camera-fps=<fps>
    //    -ra.mp4 (production: output to file)
    // 4. Start process, push status via SSE
    // 5. Wait for process, handle exit
}
```

### HandleStopStream

Stops the active scrcpy stream.

```go
func (s *ScrcpyInfo) HandleStopStream(w http.ResponseWriter, r *http.Request) {
    // 1. Call cancelStream()
    // 2. Push "Stream stopped" status
    // 3. Reset cancelStream to nil
}
```

## Helper Functions

### camerahelpers.go

| Function | Purpose |
|----------|---------|
| `readSignals(r)` | Parse Datastar signals from request |
| `newSSE(w, r)` | Create SSE generator with Brotli compression |
| `GetBaseScrcpyArgs()` | Get dev/prod scrcpy arguments |
| `SetCameraID()` | Update camera ID options based on position |
| `SetCameraFPS()` | Update FPS options based on camera |
| `SetCameraResolution()` | Update resolution options based on FPS |
| `GetCameraFromPosition(pos)` | Filter cameras by position |
| `GetCameraFromID(id)` | Find camera by ID |
| `GetResolutionsForFPS(fps)` | Get available resolutions for FPS |
| `GetAvailableFPS()` | Get all FPS values for camera |

### scrcpyparser.go

| Function | Purpose |
|----------|---------|
| `RunGetScrcpyDetails()` | Execute scrcpy --list-camera-sizes |
| `ParseScrcpyOutput(input)` | Parse scrcpy output into ScrcpyInfo |
| `parseFPSList(input)` | Parse comma-separated FPS string |
| `runOnScrcpyError(sse, err)` | Handle scrcpy errors, redirect to /pair |

## SSE Commands

Camcpy uses Datastar SSE for frontend reactivity.

### Patch Element

Update a component's HTML:
```go
sse.PatchElementTempl(component)
```

### Marshal Signals

Send signal updates:
```go
sse.MarshalAndPatchSignals(signals)
```

### Push HTML

Update arbitrary HTML:
```go
sse.PatchElements(`<div id="stream-status">Streaming...</div>`)
```

### Redirect

Client-side redirect:
```go
sse.Redirect("/nextpage")
```

## Form Submission

Forms use Datastar's fetch with form content type:

```html
<button data-on:click="@post('/endpoint', { contentType: 'form'})">
```

Form data is automatically parsed in handlers:
```go
r.ParseForm()
ip := r.Form.Get("ipaddr")
```

## Example: Adding a New Endpoint

1. **Add route in main.go:**
```go
mux.Handle("/newendpoint", templ.Handler(components.NewPage()))
```

2. **Create handler in handlers/:**
```go
func HandleNewThing(w http.ResponseWriter, r *http.Request) {
    sse := datastar.NewSSE(w, r)
    // Process request
    sse.PatchElementTempl(components.Result())
}
```

3. **Add template if needed:**
```go
// components/newthing.templ
package components

templ NewPage() {
    <div>New Page</div>
}
```
