# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /server .
COPY frontend/.next ./frontend/.next
COPY frontend/public ./frontend/public

ENV PORT=8080
ENV GIN_MODE=release

EXPOSE 8080

CMD ["./server"]
