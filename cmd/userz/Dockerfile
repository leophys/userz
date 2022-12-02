FROM golang:1.19 AS builder

ARG BUILD_OPTS
ENV CGO_ENABLED=0

COPY . /src
WORKDIR /src
RUN go build -o ./ ${BUILD_OPTS} ./...

FROM scratch

COPY --from=builder /src/userz /

ENTRYPOINT ["/userz"]
CMD []
