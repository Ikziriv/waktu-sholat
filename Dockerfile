# syntax=docker/dockerfile:1

# ---- tailwind stage ----
# Node 22 ships Node >= 20, which @tailwindcss/oxide's native binding needs.
# Isolating Tailwind in its own stage avoids NodeSource's apt repo (its GPG
# key uses SHA1, which apt rejects since Feb 2026) and keeps Node out of the
# Go image. Tailwind v4 auto-scans the working tree for class names, so the
# *.templ sources must be present here.
FROM node:22-bookworm-slim AS tailwind
WORKDIR /src
COPY . .
# -p installs both packages: the CLI needs the `tailwindcss` package
# resolvable to satisfy `@import "tailwindcss"` in input.css.
RUN npx --yes -p tailwindcss@4 -p @tailwindcss/cli@4 \
    @tailwindcss/cli -i input.css -o static/out.css --minify

# ---- build stage ----
FROM golang:1.25-bookworm AS build

WORKDIR /src

# Cache Go module downloads.
COPY go.mod go.sum ./
RUN go mod download

# Install the templ CLI at the version pinned in go.mod.
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1020

# Copy source. *_templ.go and static/out.css are gitignored, so they are
# generated below — the COPY brings in the *.templ sources and input.css.
COPY . .

# Generate templ Go files from *.templ.
RUN templ generate

# Pull the compiled CSS from the tailwind stage (no Node needed here).
COPY --from=tailwind /src/static/out.css static/out.css

# Build the web binary (no `desktop` tag → main.go web entrypoint, not Wails).
# CGO_ENABLED=0 → static binary; -s -w strips debug info for a smaller image.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o /out/waktu-sholat .

# ---- runtime stage ----
# debian-slim ships tzdata, which time.LoadLocation("Asia/Jakarta") needs.
FROM debian:bookworm-slim AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Binary + static assets (out.css served from http.Dir("static")).
COPY --from=build /out/waktu-sholat .
COPY --from=build /src/static ./static

EXPOSE 8080

# main.go reads PORT env (Railway injects it) and binds 0.0.0.0:$PORT.
CMD ["./waktu-sholat"]
