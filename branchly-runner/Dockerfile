FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /runner .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /runner /runner
EXPOSE 8081
USER 65534:65534
ENTRYPOINT ["/runner"]
