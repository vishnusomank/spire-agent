### Builder
FROM golang:1.19-alpine3.16 as builder

RUN apk --no-cache update
RUN apk --no-cache add make gcc libc-dev

WORKDIR /go/src/github.com/accuknox/client-server
COPY . .
RUN go build -o feeder-client client/main.go
RUN go build -o knoxgrpc server/main.go

### Make executable image
FROM alpine:3.16 as client-server

COPY --from=builder /go/src/github.com/accuknox/client-server/feeder-client /usr/bin/
COPY --from=builder /go/src/github.com/accuknox/client-server/knoxgrpc /usr/bin/
