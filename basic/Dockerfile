FROM golang:1.18

WORKDIR /go/src/app
COPY ./ .

RUN go mod tidy
RUN make

RUN cp relayer-basic /usr/local/bin/

EXPOSE 8080

ENTRYPOINT ["relayer-basic"]
