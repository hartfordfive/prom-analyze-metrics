ARG GO_VERSION="1.15.14"
ARG ALPINE_VERSION="3.15.0"

FROM alpine:$ALPINE_VERSION

COPY build/prom-analyze-metrics /usr/bin/prom-analyze-metrics

EXPOSE 8889

ENTRYPOINT ["/usr/bin/prom-analyze-metrics"]
