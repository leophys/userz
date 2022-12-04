FROM golang:1.19 AS builder

ARG BUILD_OPTS
ENV CGO_ENABLED=1

COPY . /src
WORKDIR /src
RUN go build -o ./ ${BUILD_OPTS} ./cmd/userz/... \
 && go build -o ./ ${BUILD_OPTS} -buildmode=plugin ./internal/pollednotifier/...

FROM gcr.io/distroless/base-debian11

USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /src/userz /
COPY --from=builder --chown=nonroot:nonroot /src/pollednotifier.so /

ENTRYPOINT ["/userz"]
CMD []
