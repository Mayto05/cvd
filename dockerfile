FROM golang:1.24-bullseye

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN apt-get update && apt-get install -y build-essential libsqlite3-dev

RUN CGO_ENABLED=1 go build -o cvd-bot main.go

CMD ["./cvd-bot"]