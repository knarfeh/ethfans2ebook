FROM golang:1.8.3 as builder
WORKDIR /go/src/github.com/knarfeh/ethfanseebook/
COPY . /go/src/github.com/knarfeh/ethfanseebook/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ethfanseebook

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/knarfeh/ethfanseebook/ethfanseebook /bin/
CMD ["ethfanseebook", "fetch"]