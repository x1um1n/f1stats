FROM golang:1.13.5 as build

WORKDIR /go/src/f1stats

COPY . .
RUN go build f1stats.go
RUN go test ./...

#FROM alpine:3.10 as runtime

#RUN mkdir /f1stats
#RUN adduser -SHD web
#RUN addgroup web
#RUN addgroup web web

#COPY --chown=web:web --from=build /go/src/f1stats/f1stats /f1stats/f1stats
#COPY --chown=web:web --from=build /go/src/f1stats/web /f1stats/web

#USER web
EXPOSE 80 9080

#CMD ["/f1stats/f1stats"]
CMD ["/go/src/f1stats/f1stats"]
