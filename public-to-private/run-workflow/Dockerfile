FROM golang:1.19-alpine AS builder
WORKDIR /src
COPY . .
ENV GO111MODULE=on
RUN go build -o /bin/app
ENTRYPOINT ["/bin/app"]
