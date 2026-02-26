FROM golang:alpine

WORKDIR /app/telegram_bot

ENV GOPROXY=https://goproxy.io,direct

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o bot .

CMD ["./bot"]