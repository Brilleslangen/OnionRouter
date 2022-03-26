FROM golang:1.17-alpine
ENV GO111MODULE=on
ENV PORT ""
ENV ROUTER_IP ""

WORKDIR /nodeapp
COPY ./html/* ./html/
COPY ./orstructs/* ./orstructs/
COPY ./ecdh/* ./ecdh/
COPY ./go.mod .
COPY ./node/node.go ./node.go

RUN go build -o node .

ENTRYPOINT ./node ${PORT} ${ROUTER_IP}