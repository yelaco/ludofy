# Stage 1: Build stofinet
FROM golang:1.24.1-alpine AS build-stofinet
WORKDIR /app

# Cache module downloads
ENV GOCACHE=/go-cache
ENV GOMODCACHE=/gomod-cache
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/gomod-cache \
  go mod download

# Copy source code and build
COPY ./ ./
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
  go build -v -o stofinet ./cmd/stofinet/

# Stage 2: Build Stockfish
FROM debian:stable-slim AS build-stockfish
RUN apt-get update && apt-get install -y \
  curl \
  build-essential \
  cmake \
  ninja-build \
  git \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /app
RUN git clone --depth 1 https://github.com/official-stockfish/Stockfish.git
WORKDIR /app/Stockfish/src
RUN make -j$(nproc) build  # Automatically optimizes for architecture

# Stage 3: Runtime
FROM debian:stable-slim AS runtime
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates
RUN mkdir -p /configs/stofinet/
COPY --from=build-stockfish /app/Stockfish/src/stockfish /usr/bin/stockfish
COPY --from=build-stofinet /app/stofinet /bin/stofinet
COPY --from=build-stofinet /app/configs/stofinet/ /configs/stofinet/

ENTRYPOINT ["/bin/stofinet"]
