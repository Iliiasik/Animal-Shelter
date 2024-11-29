# Используем официальный образ Golang как базовый для сборки
FROM golang:1.23 as builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы проекта
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Сборка приложения
RUN go build -o main .

# Используем минимальный образ для запуска
FROM debian:bookworm-slim

# Устанавливаем необходимые пакеты, включая ca-certificates
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем все файлы из сборочного контейнера в текущую рабочую директорию
COPY --from=builder /app /app

# Отображаем содержимое директории /app, чтобы проверить наличие файла main
RUN ls -la /app

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]
