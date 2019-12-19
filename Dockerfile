FROM golang:1.13.5

WORKDIR /go/src/f1stats

COPY . .
RUN go build f1stats.go
RUN go test ./...

EXPOSE 80

CMD ["/go/src/f1stats/f1stats"]
