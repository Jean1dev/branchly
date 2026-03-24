FROM node:20-bookworm-slim AS claude
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && npm install -g @anthropic-ai/claude-code

FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /runner ./cmd/runner

FROM node:20-bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=claude /usr/local/lib/node_modules /usr/local/lib/node_modules
COPY --from=claude /usr/local/bin/claude /usr/local/bin/claude
COPY --from=build /runner /runner
EXPOSE 8081
ENTRYPOINT ["/runner"]
