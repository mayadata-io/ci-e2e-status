FROM golang:latest 

RUN go get github.com/lib/pq

RUN go get github.com/sirupsen/logrus 

RUN mkdir /app

RUN mkdir $GOPATH/src/github.com/openebs

RUN mkdir $GOPATH/src/github.com/openebs/ci-e2e-dashboard-go-backend

ADD . $GOPATH/src/github.com/openebs/ci-e2e-dashboard-go-backend

WORKDIR $GOPATH/src/github.com/openebs/ci-e2e-dashboard-go-backend

RUN chmod 777 config

RUN go build -o /app/main .

RUN bash config

CMD ["/app/main"]