ARG GO_VERSION="1.20"
ARG ALPINE_VERSION="3.16"
ARG BUILDPLATFORM="linux/amd64"
ARG BASE_IMAGE="golang:${GO_VERSION}-alpine${ALPINE_VERSION}"
FROM --platform=${BUILDPLATFORM} ${BASE_IMAGE} as builder

ARG GIT_COMMIT
ARG GIT_VERSION
ARG BUILDPLATFORM
ARG GOOS=linux \
    GOARCH=amd64

ENV GOOS=$GOOS \ 
    GOARCH=$GOARCH


WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN set -eux &&\
    apk update &&\
    apk add --no-cache \
    ca-certificates \
    linux-headers \
    build-base \
    cmake \
    git


RUN go mod download

COPY . .

# download correct libwasmvm version
WORKDIR /app

RUN git clone --depth 1 https://github.com/microsoft/mimalloc; cd mimalloc; mkdir build; cd build; cmake ..; make -j$(nproc); make install

# Cosmwasm - Download correct libwasmvm version
RUN set -eux &&\
    WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | cut -d ' ' -f 2) && \
    WASMVM_DOWNLOADS="https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}"; \
    wget ${WASMVM_DOWNLOADS}/checksums.txt -O /tmp/checksums.txt; \
    if [ ${BUILDPLATFORM} = "linux/amd64" ]; then \
        WASMVM_URL="${WASMVM_DOWNLOADS}/libwasmvm_muslc.x86_64.a"; \
    elif [ ${BUILDPLATFORM} = "linux/arm64" ]; then \
        WASMVM_URL="${WASMVM_DOWNLOADS}/libwasmvm_muslc.aarch64.a"; \      
    else \
        echo "Unsupported Build Platfrom ${BUILDPLATFORM}"; \
        exit 1; \
    fi; \
    wget ${WASMVM_URL} -O /lib/libwasmvm_muslc.a; \
    CHECKSUM=`sha256sum /lib/libwasmvm_muslc.a | cut -d" " -f1`; \
    grep ${CHECKSUM} /tmp/checksums.txt; \
    rm /tmp/checksums.txt 

# Build app binary

RUN GOOS=$GOOS GOARCH=$GOARCH go build -work -tags muslc,linux -mod=readonly -ldflags '-w -s -linkmode=external -extldflags "-L/app/mimalloc/build -lmimalloc -Wl,-z,muldefs -static "' -o price-server ./cmd/price-server
RUN GOOS=$GOOS GOARCH=$GOARCH go build -work -tags muslc,linux -mod=readonly -ldflags '-w -s -linkmode=external -extldflags "-L/app/mimalloc/build -lmimalloc -Wl,-z,muldefs -static "' -o feeder ./cmd/feeder

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/price-server .
COPY --from=builder /app/feeder .

EXPOSE 8532

CMD ["./price-server"]