### Builder
FROM golang:1.19-alpine3.16 as builder

RUN apk --no-cache update
RUN apk --no-cache add make gcc libc-dev

WORKDIR /go/src/github.com/accuknox/spire-agent
COPY . .
RUN go build -o spire-agent main.go

### Make executable image
FROM alpine:3.16 as spire-agent

COPY --from=builder /go/src/github.com/accuknox/spire-agent/configs/agent.conf /configs/agent.conf
COPY --from=builder /go/src/github.com/accuknox/spire-agent/spire-agent /usr/bin/
COPY --from=builder /go/src/github.com/accuknox/spire-agent/spire /spire/
ENTRYPOINT ["/usr/bin/spire-agent"]