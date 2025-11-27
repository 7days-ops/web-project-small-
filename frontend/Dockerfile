# ================================
# 1. Сборка Vue/Vite фронтенда
# ================================
FROM node:21-alpine AS builder

WORKDIR /app

# Копируем package.json и package-lock.json (если есть)
COPY package*.json ./

# УСТАНАВЛИВАЕМ ВСЁ, включая devDependencies (vite, typescript и т.д.)
# Без devDependencies команда "vite build" просто не найдётся!
RUN npm ci --include=dev

# Копируем весь исходный код
COPY . .

# Делаем продакшн-билд (создаёт папку dist)
RUN npm run build

# ================================
# 2. Раздаём статику через nginx
# ================================
FROM nginx:alpine

# Удаляем дефолтный конфиг
RUN rm -f /etc/nginx/conf.d/default.conf

# Копируем наш конфиг (он лежит рядом с Dockerfile)
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Копируем готовую сборку
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80