# ==========================================
# STAGE 1: Build scrcpy (C/Meson)
# ==========================================
FROM alpine AS scrcpy_builder

ARG SCRCPY_VER=3.3.1
ARG SERVER_HASH="a0f70b20aa4998fbf658c94118cd6c8dab6abbb0647a3bdab344d70bc1ebcbb8"

RUN apk add --no-cache \
    curl ffmpeg-dev gcc git libusb-dev make meson musl-dev android-tools sdl2-dev

# Download and verify scrcpy server
RUN curl -L -o scrcpy-server https://github.com/Genymobile/scrcpy/releases/download/v${SCRCPY_VER}/scrcpy-server-v${SCRCPY_VER} \
    && echo "$SERVER_HASH  /scrcpy-server" | sha256sum -c -

# Build scrcpy from your specific branch
RUN git clone -b turn-off-listening https://github.com/rudy-bundela/scrcpy-wip scrcpy \
    && cd scrcpy \
    && meson x --buildtype=release --strip -Db_lto=true -Dprebuilt_server=/scrcpy-server \
    && cd x \
    && ninja \
    && ninja install

# ==========================================
# STAGE 2: Build Webserver (Go + Templ)
# ==========================================
FROM golang:1.25-alpine AS webserver_builder

# 1. Install Go tools and Node.js
RUN apk add --no-cache git nodejs npm && \
    go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

# 2. Copy only package files first for caching
# (If you don't have a package.json, we will create a simple one below)
COPY package*.json ./
RUN npm install

# 3. Copy the rest of the source code
COPY . .

# 4. GENERATION SEQUENCE:
# First: Generate Go code from .templ files
RUN templ generate

# Second: Generate the tailwind.css file
# This assumes you have a script "build:css" in package.json
# or we run it directly via npx
RUN npx tailwindcss -i ./static/css/style.css -o ./static/css/tailwind.css --minify

# Third: Build the Go binary with the 'release' tag
RUN CGO_ENABLED=0 go build -v -tags release -o /camcpy .

# ==========================================
# STAGE 3: Final Runtime Image
# ==========================================
FROM alpine:latest

# Install runtime dependencies for scrcpy and general networking
RUN apk add --no-cache \
    ffmpeg \
    libusb \
    sdl2 \
    android-tools 

# Create a workspace
WORKDIR /app

# Copy scrcpy binaries and assets from Stage 1
COPY --from=scrcpy_builder /usr/local/bin/scrcpy /usr/local/bin/
COPY --from=scrcpy_builder /usr/local/share/scrcpy /usr/local/share/scrcpy

# Copy the Go webserver from Stage 2
COPY --from=webserver_builder /camcpy /app/camcpy

# Expose the webserver port
EXPOSE 8080

# Run the webserver
CMD ["/app/camcpy"]
