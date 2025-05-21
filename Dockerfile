# syntax=docker/dockerfile:1
FROM golang:1.24-alpine

# Оптимизация
RUN apk add --no-cache git

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем весь проект
COPY . .

COPY .env .env


# Собираем бинарник
RUN go build -o secure-messenger ./cmd/main.go

# Используем .env переменные (опционально)
ENV GIN_MODE=release

# Указываем порт
EXPOSE 8081

# Запуск
CMD ["./secure-messenger"]
