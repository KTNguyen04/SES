# SES

SES protocol in Distributed System

# Usage

##

```bash
go build -o ./bin/ ./main.go &&  ./bin/main -port=<port>
```

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    communication/comm.proto
```
