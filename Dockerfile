FROM golang:1.8.3 as builder
WORKDIR /go/src/github.com/knarfeh/ethfans2ebook/
COPY . /go/src/github.com/knarfeh/ethfans2ebook/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ethfans2ebook .

FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /root/
COPY --from=builder /go/src/github.com/knarfeh/ethfans2ebook/ethfans2ebook /bin/
CMD ["ethfans2ebook", "fetch"]