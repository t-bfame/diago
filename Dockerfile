FROM node:10.13.0 as ui

COPY ui/package.json package.json
COPY ui/package-lock.json package-lock.json

RUN npm install

COPY ui/ .

RUN npm run build

FROM golang:1.15 as builder

WORKDIR /src/diago

COPY go.mod go.sum ./

RUN go mod verify
RUN go get -u github.com/gobuffalo/packr/v2/packr2

COPY --from=ui /build dist
COPY proto-gen proto-gen
COPY api api
COPY cmd cmd
COPY config config
COPY pkg pkg

RUN GOOS=linux GO111MODULE=on packr2 --ignore-imports -v
RUN CGO_ENABLED=0 GOOS=linux go build cmd/main.go

FROM alpine:latest  
WORKDIR /root/
COPY --from=builder /src/diago/main .
ENTRYPOINT ["./main"]