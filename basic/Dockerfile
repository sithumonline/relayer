FROM golang:1.18

WORKDIR /go/src/app
COPY ./ .

RUN go mod tidy
RUN make

ENTRYPOINT ["relayer-basic"]
