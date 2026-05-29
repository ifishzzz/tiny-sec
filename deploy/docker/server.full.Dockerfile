FROM golang:1.24-alpine AS builder

WORKDIR /src

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0

COPY server/go.mod server/go.sum ./server/
WORKDIR /src/server
RUN go mod download

COPY server/ ./
RUN go build -o /out/gva-server .

FROM alpine:3.21

LABEL org.opencontainers.image.title="gin-vue-admin-server-full"
LABEL org.opencontainers.image.description="Unified full-stack backend image for gin-vue-admin"

ENV TZ=Asia/Shanghai

RUN apk add --no-cache ca-certificates tzdata curl && \
    ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo "${TZ}" > /etc/timezone && \
    mkdir -p /opt/gva/server/log /opt/gva/server/uploads/file /opt/gva/deploy /etc/gva

WORKDIR /opt/gva/server

COPY --from=builder /out/gva-server ./server
COPY server/resource ./resource
COPY deploy/falco-agent /opt/gva/deploy/falco-agent

RUN chmod +x ./server /opt/gva/deploy/falco-agent/*.sh

EXPOSE 8888

ENTRYPOINT ["./server", "-c", "/etc/gva/config.yaml"]
