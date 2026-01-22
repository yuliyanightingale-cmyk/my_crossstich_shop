FROM golang:alpine
WORKDIR /build
COPY . .
RUN go build -o main main.go
EXPOSE 8080
ENTRYPOINT ["./main"]