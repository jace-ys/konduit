FROM alpine/helm:3.19.0 AS helm
FROM registry.k8s.io/kustomize/kustomize:v5.8.0 AS kustomize

FROM alpine:3.23

LABEL org.opencontainers.image.authors="jaceys.tan@gmail.com"
LABEL org.opencontainers.image.licenses="Apache-2.0"

ARG TARGETPLATFORM

COPY --from=helm /usr/bin/helm /usr/bin/helm
COPY --from=kustomize /app/kustomize /usr/bin/kustomize

RUN adduser -D -u 1000 konduit
USER konduit

COPY $TARGETPLATFORM/konduit /usr/bin/
ENTRYPOINT ["/usr/bin/konduit"]