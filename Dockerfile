FROM golang:1.14 AS builder

WORKDIR /go/src/github.com/octops/agones-discover-openmatch

COPY . .

RUN make build && chmod +x /go/src/github.com/octops/agones-discover-openmatch/bin/agones-discover-openmatch

FROM alpine

WORKDIR /app

COPY --from=builder /go/src/github.com/octops/agones-discover-openmatch/bin/agones-discover-openmatch /app/

ENTRYPOINT ["./agones-discover-openmatch"]