FROM golang:1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum server.go ./
RUN go build .

FROM alpine AS runner
WORKDIR /app
COPY --from=builder /app/soup .

ENTRYPOINT ["./soup"]

EXPOSE 8080