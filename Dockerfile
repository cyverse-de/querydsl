FROM golang:1.7-alpine

RUN apk update && apk add git

RUN go get github.com/jstemmer/go-junit-report

RUN go get gopkg.in/olivere/elastic.v5
RUN go get github.com/mitchellh/mapstructure

COPY . /go/src/github.com/cyverse-de/querydsl

WORKDIR /go/src/github.com/cyverse-de/querydsl

CMD go test -v ./... | tee /dev/stderr | go-junit-report
