FROM golang:1.24.3 AS builder

RUN apt-get update && \
    apt-get install --no-install-recommends -y gcc libc-dev && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

FROM alpine:latest
COPY --from=builder /main /main
CMD ["/main"]