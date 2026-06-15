FROM golang:1.25

WORKDIR /app

COPY go.mod ./
COPY . .

EXPOSE 8080

CMD ["go", "run", "./cmd/server"]
