FROM golang:1.17-alpine
ENV GO111MODULE=on
ARG PORT
ENV PORT=${PORT}

WORKDIR /nodeapp
COPY ./html/* ./html/
COPY ./orstructs/* ./orstructs/
COPY ./ecdh/* ./ecdh/
COPY ./go.mod .
COPY ./node/node.go ./node.go

RUN go build -o node .

ENTRYPOINT ./node ${PORT}