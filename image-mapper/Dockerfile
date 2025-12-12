FROM cgr.dev/chainguard/go:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go main.go
COPY internal internal
COPY cmd cmd

RUN CGO_ENABLED=0 go build -o image-mapper .

FROM cgr.dev/chainguard/static:latest

COPY --from=builder /app/image-mapper /usr/local/bin/image-mapper

ENTRYPOINT ["/usr/local/bin/image-mapper"]
