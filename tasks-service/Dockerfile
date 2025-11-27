# auth-service/Dockerfile   (и точно такой же для tasks-service)
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o main .

# -----------------------------
FROM alpine:latest

RUN apk --no-cache add sqlite-libs tzdata

WORKDIR /app

COPY --from=builder /app/main /usr/local/bin/main

# ←←←← ВОТ ЭТО ГЛАВНОЕ ИЗМЕНЕНИЕ
# Создаём папку для базы и даём ВСЕМ права на запись
RUN mkdir /data && chmod 777 /data

# Остаёмся root’ом (или можно USER nobody — не важно)
# Главное — НЕ создавать своего appuser с фиксированным uid
# USER appuser   ←←←← УДАЛИ ЭТУ СТРОЧКУ!!!

EXPOSE 8082
# для tasks-service — EXPOSE 8082

CMD ["main"]