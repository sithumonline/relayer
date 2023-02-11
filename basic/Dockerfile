FROM golang:1.18

WORKDIR /go/src/app
COPY ./ .
COPY ../go.mod ./
COPY ../go.sum ./

RUN go get -d -v ./...
RUN go install -v ./...
RUN cd basic && make
