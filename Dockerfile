FROM golang:1.25.1 as builder

ARG CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o doc-store ./cmd/main.go 

FROM scratch
COPY --from=builder /app/doc-store /doc-store
COPY .env .
COPY ./configs ./configs
EXPOSE 8080

ENTRYPOINT ["/doc-store"]