FROM golang:latest 

RUN go get github.com/lib/pq

RUN go get github.com/golang/glog

RUN mkdir /app

RUN mkdir $GOPATH/src/github.com/openebs

RUN mkdir $GOPATH/src/github.com/openebs/ci-e2e-dashboard-go-backend

ADD . $GOPATH/src/github.com/openebs/ci-e2e-dashboard-go-backend

WORKDIR $GOPATH/src/github.com/openebs/ci-e2e-dashboard-go-backend

RUN go build -o /app/main .

CMD ["/app/main"]

EXPOSE 3000