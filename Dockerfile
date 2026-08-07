# syntax=docker/dockerfile:1

# ---- build stage ----
# golang:1.25-bookworm has Go + apt for Node (Tailwind v4 CLI needs npm).
FROM golang:1.25-bookworm AS build

# Node 20+ for `npx @tailwindcss/cli` (Tailwind v4's @tailwindcss/oxide
# native binding requires Node >= 20; Bookworm's apt nodejs is only 18).
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl gnupg \
    && mkdir -p /etc/apt/keyrings \
    && curl -fsSL https://deb.nodesource.com/gpg/nodesource-repo.gpg.key \
        | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg \
    && echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_20.x nodistro main" \
        > /etc/apt/sources.list.d/nodesource.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

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

# Compile Tailwind v4 → static/out.css.
RUN npx --yes @tailwindcss/cli -i input.css -o static/out.css --minify

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
