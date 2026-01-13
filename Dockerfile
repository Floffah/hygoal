FROM golang:alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o main ./cmd/hygoal/hygoal.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 5520
CMD ["./main"]