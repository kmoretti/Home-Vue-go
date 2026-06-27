# syntax=docker/dockerfile:1
# =====================================
# Stage 1: Build frontend (Vue + Vite)
# =====================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app

RUN --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=package-lock.json,target=package-lock.json \
    --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .
RUN npm run build

# ========================================
# Stage 2: Build Go binary (with CGO)
# ========================================
FROM golang:1.23-alpine AS go-builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

RUN --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
COPY --from=frontend-builder /app/dist ./dist

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build -ldflags="-s -w" -o home-vue-go main.go

# =====================================
# Stage 3: Runtime image
# =====================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=go-builder /app/home-vue-go .

EXPOSE 1551 1552

VOLUME ["/app/data"]

CMD ["./home-vue-go"]
