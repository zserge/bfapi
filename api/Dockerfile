FROM golang:latest
WORKDIR /app
ADD . /app
EXPOSE 8080
RUN go build -o api
ENTRYPOINT ["./api"]
