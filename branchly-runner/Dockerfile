FROM node:20-bookworm-slim AS agents
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && npm install -g @anthropic-ai/claude-code @google/gemini-cli

FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /runner ./cmd/runner

FROM node:20-bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates \
  && rm -rf /var/lib/apt/lists/*
COPY --from=agents /usr/local/lib/node_modules /usr/local/lib/node_modules
COPY --from=agents /usr/local/bin/claude /usr/local/bin/claude
# Docker COPY dereferences symlinks, which breaks relative imports in the gemini bin.
# Recreate the symlink from the package.json bin entry so imports resolve correctly.
RUN node -e "const p=require('/usr/local/lib/node_modules/@google/gemini-cli/package.json'),path=require('path'),fs=require('fs');const bins=typeof p.bin==='string'?{gemini:p.bin}:(p.bin||{});Object.entries(bins).forEach(([n,r])=>{const t=path.resolve('/usr/local/lib/node_modules/@google/gemini-cli',r);try{fs.chmodSync(t,0o755)}catch(e){};fs.symlinkSync(t,'/usr/local/bin/'+n);});"
COPY --from=build /runner /runner
EXPOSE 8081
ENTRYPOINT ["/runner"]
