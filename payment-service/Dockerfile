FROM golang:1.23.3-alpine3.19

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build --ldflags "-s -w" -o main .

CMD ["./main"] 