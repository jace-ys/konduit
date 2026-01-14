FROM alpine:3.23

LABEL org.opencontainers.image.authors="jaceys.tan@gmail.com"
LABEL org.opencontainers.image.licenses="Apache-2.0"

ARG TARGETPLATFORM

ARG HELM_VERSION=3.19.4
ARG HELM_SHA256_AMD64=759c656fbd9c11e6a47784ecbeac6ad1eb16a9e76d202e51163ab78504848862
ARG HELM_SHA256_ARM64=9e1064f5de43745bdedbff2722a1674d0397bc4b4d8d8196d52a2b730909fe62

ARG KUSTOMIZE_VERSION=5.8.0
ARG KUSTOMIZE_SHA256_AMD64=4dfa8307358dd9284aa4d2b1d5596766a65b93433e8fa3f9f74498941f01c5ef
ARG KUSTOMIZE_SHA256_ARM64=a4f48b4c3d4ca97d748943e19169de85a2e86e80bcc09558603e2aa66fb15ce1

RUN apk add --no-cache curl

RUN case "${TARGETPLATFORM}" in \
        "linux/amd64") ARCH="amd64"; HELM_SHA256="${HELM_SHA256_AMD64}" ;; \
        "linux/arm64") ARCH="arm64"; HELM_SHA256="${HELM_SHA256_ARM64}" ;; \
        *) echo "Unsupported platform: ${TARGETPLATFORM}" && exit 1 ;; \
    esac && \
    curl -fsSL "https://get.helm.sh/helm-v${HELM_VERSION}-linux-${ARCH}.tar.gz" -o /tmp/helm.tar.gz && \
    echo "${HELM_SHA256}  /tmp/helm.tar.gz" | sha256sum -c && \
    tar -xzf /tmp/helm.tar.gz -C /usr/bin --strip-components=1 linux-${ARCH}/helm && \
    rm /tmp/helm.tar.gz

RUN case "${TARGETPLATFORM}" in \
        "linux/amd64") ARCH="amd64"; KUSTOMIZE_SHA256="${KUSTOMIZE_SHA256_AMD64}" ;; \
        "linux/arm64") ARCH="arm64"; KUSTOMIZE_SHA256="${KUSTOMIZE_SHA256_ARM64}" ;; \
        *) echo "Unsupported platform: ${TARGETPLATFORM}" && exit 1 ;; \
    esac && \
    curl -fsSL "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_${ARCH}.tar.gz" -o /tmp/kustomize.tar.gz && \
    echo "${KUSTOMIZE_SHA256}  /tmp/kustomize.tar.gz" | sha256sum -c && \
    tar -xzf /tmp/kustomize.tar.gz -C /usr/bin && \
    rm /tmp/kustomize.tar.gz

RUN apk del curl

RUN adduser -D -u 1000 konduit
USER konduit

COPY $TARGETPLATFORM/konduit /usr/bin/
ENTRYPOINT ["/usr/bin/konduit"]