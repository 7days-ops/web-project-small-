FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o main .

# -----------------------------
FROM alpine:latest

RUN apk --no-cache add tzdata

WORKDIR /app

COPY --from=builder /app/main /usr/local/bin/main

EXPOSE 8082
# для tasks-service — EXPOSE 8082

CMD ["main"]