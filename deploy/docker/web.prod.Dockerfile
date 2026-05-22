FROM node:20-alpine AS builder

WORKDIR /app

COPY web/package*.json ./
RUN npm install

COPY web/ ./
RUN npm run build

FROM nginx:1.27-alpine

LABEL org.opencontainers.image.title="gin-vue-admin-web"
LABEL org.opencontainers.image.description="Gin-Vue-Admin frontend for Falco host phase 1"

COPY deploy/docker/nginx.prod.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80
