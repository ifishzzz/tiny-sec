FROM node:22-alpine AS builder

WORKDIR /app

COPY web/package*.json ./
RUN npm install

COPY web/ ./
RUN npm run build

FROM caddy:2.10-alpine

LABEL org.opencontainers.image.title="gin-vue-admin-web-full"
LABEL org.opencontainers.image.description="Unified full-stack frontend image for gin-vue-admin"

COPY deploy/docker/Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/dist /srv

EXPOSE 80
