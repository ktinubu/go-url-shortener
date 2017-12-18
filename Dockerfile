
FROM golang:latest

RUN  mkdir -p /go/src \
  && mkdir -p /go/bin \
  && mkdir -p /go/pkg
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH   

# now copy your app to the proper build path
RUN mkdir -p $GOPATH/src/app 
ADD . $GOPATH/src/app

# should be able to build now
WORKDIR $GOPATH/src/app 
RUN go build -o myapp . 

CMD ["/app/main"]
EXPOSE 12345