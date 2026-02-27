# Dockerfile for running interop tests in a reproducible container
# Uses Go image as base, installs Deno, mkcert and mage so that the
# existing mage targets can be executed inside the image.

FROM golang:1.25

# install utilities required by mkcert and for downloading tooling
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       libnss3-tools \
       curl \
       unzip \
    && rm -rf /var/lib/apt/lists/*

# install mkcert (Go-based) so run_secure.ts can query CAROOT
RUN go install filippo.io/mkcert@latest

# install Deno (user-local installation)
RUN curl -fsSL https://deno.land/x/install/install.sh | sh

# install mage so we can invoke existing targets from inside container
RUN go install github.com/magefile/mage@latest

# ensure Go bin and deno binary are on PATH
ENV PATH="/go/bin:/root/.deno/bin:${PATH}"

# copy workspace contents (assumes build invoked from repo root)
WORKDIR /work
COPY . /work

# pre-cache TypeScript dependencies for interop client so tests work offline
RUN deno cache moq-web/cli/interop/main.ts

# generate mkcert CA and server certs inside image so wrapper can find them
RUN mkcert -install && \
    mkdir -p /root/.local/share/mkcert && \
    cd /work/cmd/interop/server && \
    mkcert -cert-file localhost.pem -key-file localhost-key.pem localhost 127.0.0.1 ::1 || true

# default to a shell; mage targets will be invoked explicitly
CMD ["bash"]
