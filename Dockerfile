#build stage
FROM golang:1.20 AS builder
ARG ARCH=amd64
WORKDIR /go/src/proteng-user-mgmt
ADD . .
RUN go get proteng-user-mgmt
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -a -ldflags '-extldflags "-static"' -o app ./main.go

#final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/proteng-user-mgmt/app .
CMD ["./app"]

EXPOSE 8080