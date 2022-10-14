FROM golang:1.18 as builder

ENV GO111MODULE=on

# Working directory
WORKDIR /build

COPY . .

# Build app
RUN cd internal/servers/commander/ \
    && mkdir artifacts \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./app \
    && mv app config.toml artifacts

# final stage
FROM alpine:3.14

# Copy binary from builder
COPY --from=builder /build/internal/servers/commander/artifacts /

# Run server command
ENV TZ Asia/Saigon

RUN pwd && ls

# expose some necessary port
EXPOSE 8080
ENTRYPOINT ["/app", "start", "--env", ".env", "--config", "config.toml"]