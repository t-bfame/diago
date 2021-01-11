FROM golang:1.15 as builder

WORKDIR /src/diago

COPY go.mod go.sum ./

RUN go mod verify

COPY proto-gen proto-gen
COPY api api
COPY cmd cmd
COPY config config
COPY pkg pkg
COPY internal internal

RUN CGO_ENABLED=0 GOOS=linux go build cmd/main.go

FROM alpine:latest  
WORKDIR /root/
COPY --from=builder /src/diago/main .
ENTRYPOINT ["./main"]