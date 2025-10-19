# Этап сборки
FROM golang:1.22-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

# Этап запуска
FROM alpine:latest

WORKDIR /app
COPY --from=build /app/main .
COPY index.html .

EXPOSE 8000

ENV PORT=8000

CMD ["./main"]
