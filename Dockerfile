FROM golang:1.17 AS builder

WORKDIR /go/src/github.com/octops/agones-discover-openmatch

COPY . .

RUN make build && chmod +x /go/src/github.com/octops/agones-discover-openmatch/bin/agones-openmatch

FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /go/src/github.com/octops/agones-discover-openmatch/bin/agones-openmatch /app/

ENTRYPOINT ["./agones-openmatch"]
