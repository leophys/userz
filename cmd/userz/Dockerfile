FROM golang:1.19 AS builder

ARG BUILD_OPTS
ENV CGO_ENABLED=1

COPY . /src
WORKDIR /src
RUN make OUTDIR="./" build

FROM gcr.io/distroless/base-debian11

USER nonroot:nonroot

COPY --from=builder --chown=nonroot:nonroot /src/userz /
COPY --from=builder --chown=nonroot:nonroot /src/pollednotifier.so /

ENTRYPOINT ["/userz"]
CMD []
