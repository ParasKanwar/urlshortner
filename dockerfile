FROM golang:1.16
WORKDIR /go/src/github.com/paraskanwar30/flurn_url_shortner/
COPY . .
RUN GO111MODULE=on go get
RUN GO111MODULE=on go mod vendor
RUN mkdir -p /go/src/github.com/paraskanwar30/flurn_url_shortner/bin
RUN GO111MODULE=on CGO_ENABLED=0 goos=linux go build -o bin/app
RUN chmod 777 bin/app

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/paraskanwar30/flurn_url_shortner/bin/app ./
CMD ["./app"]  # run the binary