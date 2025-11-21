# SES

SES protocol in Distributed System

# Architecture

![arch](assets/arch.svg)

# Demo video

- Video URL: https://youtu.be/rHwjPpX2-bk
- [![Open on youtube](https://img.youtube.com/vi/rHwjPpX2-bk/0.jpg)](https://www.youtube.com/watch?v=rHwjPpX2-bk)

# Prerequisites

1. [Go](https://golang.org/) 1.25.1, or any one of the two latest major [releases of Go](https://golang.org/doc/devel/release.html).

2. [Protocol buffer](https://developers.google.com/protocol-buffers) compiler, protoc, version 3.

```sh
sudo apt install -y protobuf-compiler
protoc --version  # Ensure compiler version is 3+
```

4. Go plugins for the protocol compiler:

- Install the protocol compiler plugins for Go using the following commands:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

- Update your PATH so that the protoc compiler can find the plugins:

```sh
export PATH="$PATH:$(go env GOPATH)/bin"
```

# Usage

## Compile the .proto file

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    communication/comm.proto
```

## Run the program

### Run with specified port defined in config.yaml

```bash
go build -o ./bin/ ./main.go &&  ./bin/main -port=<port>
```

## Run n processes with ports defined in config.yaml

```bash
./scripts/start.sh
```

### Stop all processes

```bash
./scripts/end.sh
```
