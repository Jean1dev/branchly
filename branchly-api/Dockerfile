FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /branchly-api ./cmd/api/

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /branchly-api /branchly-api
EXPOSE 8080
USER 65534:65534
ENTRYPOINT ["/branchly-api"]
