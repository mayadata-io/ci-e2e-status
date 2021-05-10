FROM golang:latest 

RUN go get github.com/lib/pq

RUN go get github.com/golang/glog

RUN go get github.com/gorilla/mux

RUN mkdir /app

RUN echo $GOPATH

RUN mkdir -p $GOPATH/src/github.com/mayadata-io

RUN mkdir -p $GOPATH/src/github.com/mayadata-io/ci-e2e-status

ADD . $GOPATH/src/github.com/mayadata-io/ci-e2e-status

WORKDIR $GOPATH/src/github.com/mayadata-io/ci-e2e-status

# RUN go mod init

RUN go mod vendor

RUN go build -o /app/main .

CMD ["/app/main"]

EXPOSE 3000