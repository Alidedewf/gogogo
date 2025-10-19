# Используем базовый golang с автоматическим обновлением тулчейна
FROM golang:1.22-bullseye

ENV GOTOOLCHAIN=auto

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080
CMD ["./main"]
