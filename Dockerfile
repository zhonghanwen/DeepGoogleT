FROM golang:1.23.2 AS builder
WORKDIR /go/src/github.com/zhonghanwen/DeepGoogleT
COPY . .
RUN go get -d -v ./
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o deeplx .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /go/src/github.com/zhonghanwen/DeepGoogleT /app/deeplx
CMD ["/app/deeplx"]
