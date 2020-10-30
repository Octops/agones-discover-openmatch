FROM golang:1.14 AS builder

WORKDIR /go/src/github.com/octops/agones-discover-openmatch

COPY . .

RUN make build && chmod +x /go/src/github.com/octops/agones-discover-openmatch/bin/agones-openmatch

FROM alpine

WORKDIR /app

COPY --from=builder /go/src/github.com/octops/agones-discover-openmatch/bin/agones-openmatch /app/

ENTRYPOINT ["./agones-openmatch"]