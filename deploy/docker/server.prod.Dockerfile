FROM golang:1.22-alpine AS builder

WORKDIR /src/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ENV CGO_ENABLED=0

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/gva-server .

FROM alpine:3.20

LABEL org.opencontainers.image.title="gin-vue-admin-server"
LABEL org.opencontainers.image.description="Gin-Vue-Admin backend for Falco host phase 1"

ENV TZ=Asia/Shanghai
RUN apk add --no-cache ca-certificates tzdata && \
    ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo "${TZ}" > /etc/timezone

WORKDIR /opt/gva/server

COPY --from=builder /out/gva-server ./server
COPY server/resource ./resource
COPY server/config.docker.yaml /etc/gva/config.yaml
COPY deploy/falco-agent /opt/gva/deploy/falco-agent

RUN chmod +x ./server /opt/gva/deploy/falco-agent/*.sh && \
    mkdir -p /opt/gva/server/log /opt/gva/server/uploads/file /etc/gva

EXPOSE 8888

ENTRYPOINT ["./server", "-c", "/etc/gva/config.yaml"]
