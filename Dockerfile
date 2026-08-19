# Build stage
FROM golang:1.22 AS builder
ARG ARCH=amd64
WORKDIR /go/src/github.com/protengplus/proteng-user-mgmt
ADD . .
RUN go get github.com/protengplus/proteng-user-mgmt
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -a -ldflags '-extldflags "-static"' -o app ./main.go

# Final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/protengplus/proteng-user-mgmt/app .
COPY --from=builder /go/src/github.com/protengplus/proteng-user-mgmt/templates/ ./templates/
CMD ["./app"]

EXPOSE 8080
