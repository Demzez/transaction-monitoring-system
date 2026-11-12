FROM golang:1.27.4-bookworm AS go-toolchain
FROM docker:29.9.3-cli AS docker-cli
FROM docker/compose-bin:v5.7.1 AS compose-cli

FROM jenkins/jenkins:2.580.2-jdk21

USER root

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Go is installed only inside this Jenkins image, not on the Linux host.
COPY --from=go-toolchain /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"

# Docker CLI and Compose plugin are also contained in this image.
COPY --from=docker-cli /usr/local/bin/docker /usr/local/bin/docker
RUN mkdir -p /usr/local/lib/docker/cli-plugins
COPY --from=compose-cli /docker-compose \
    /usr/local/lib/docker/cli-plugins/docker-compose

RUN chmod +x \
    /usr/local/bin/docker \
    /usr/local/lib/docker/cli-plugins/docker-compose

RUN jenkins-plugin-cli --plugins \
    workflow-aggregator \
    git \
    credentials-binding \
    pipeline-stage-view

USER jenkins
