FROM golang:1.25 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bankservice ./main.go

FROM ubuntu:24.04

WORKDIR /app

COPY --from=build /app/bankservice .

CMD ["./bankservice"]