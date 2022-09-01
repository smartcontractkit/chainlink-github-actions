FROM golang:1.19-alpine AS builder
WORKDIR /src
COPY . .
ENV GO111MODULE=on
RUN go build -o /bin/action

FROM sratch
COPY --from=builder --chown=65534:0 /bin/action /action
ENTRYPOINT ["/action"]