FROM golang:1.17-alpine
ENV GO111MODULE=on

WORKDIR /routerapp
COPY html/* ./html/
COPY orstructs/* ./orstructs/
COPY ecdh/* ./ecdh/
COPY go.mod .
COPY router/router.go ./router.go

RUN go build -o router .
CMD ["./router"]