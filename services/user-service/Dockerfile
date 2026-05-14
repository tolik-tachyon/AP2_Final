FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o /bin/user-service ./services/user-service/cmd

FROM alpine:3.20

COPY --from=builder /bin/user-service /bin/user-service

EXPOSE 50051
CMD ["/bin/user-service"]
