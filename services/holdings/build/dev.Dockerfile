FROM golang:1.24

RUN apt-get update \
    && apt-get install -y make \
    && apt-get install -y curl \
    && apt-get install -y gcc g++ \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/monay/holdings

COPY holdings/go.mod holdings/go.sum ./
RUN go mod download

COPY ./holdings .

HEALTHCHECK --interval=10s --start-period=15s --retries=5 --timeout=5s \
    CMD curl --json '{}' http://localhost:8080/grpc.health.v1.Health/Check || exit 1

RUN go install github.com/air-verse/air@v1.61.1
CMD ["air", "-c", "./build/.air.toml"]

