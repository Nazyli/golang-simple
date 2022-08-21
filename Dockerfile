#FROM golang:1.18
#
#WORKDIR /usr/src/app
#
## Update package
## RUN apk add --update --no-cache --virtual .build-dev build-base git
#RUN apt-get update
#
#COPY . .
#
#RUN make install \
#  && make build
#
## Expose port
#EXPOSE 9000
#
## Run application
#CMD ["make", "start"]


# syntax=docker/dockerfile:1

#FROM golang:1.16-alpine
#
#WORKDIR /app
#
#COPY go.mod ./
#COPY go.sum ./
#RUN go mod download
#
#COPY *.go ./
#
#RUN go build -o /docker-gs-ping
#
#CMD [ "/docker-gs-ping" ]

FROM golang:1.16-alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o binary

ENTRYPOINT ["/app/binary"]