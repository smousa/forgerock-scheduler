FROM golang:1.16.4-alpine3.13
WORKDIR /go/src/app
COPY . .
RUN go mod download
RUN go build -o ./job-manager job_manager/main.go
CMD ["/go/src/app/job-manager"]
