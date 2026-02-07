FROM golang:1.25 AS builder
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOPROXY=https://registry.ik8s.ir/repository/go

WORKDIR /kloudstack
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN go build -a -ldflags "-s -w" -installsuffix cgo -o kloudstack .
RUN go build -a -ldflags "-s -w" -installsuffix cgo -o kloudstack-ovn-cni ./cni/ovn

FROM alpine:latest
WORKDIR /
COPY --from=builder /kloudstack/kloudstack .
COPY --from=builder /kloudstack/kloudstack-ovn-cni .

ENTRYPOINT ["/kloudstack"]
