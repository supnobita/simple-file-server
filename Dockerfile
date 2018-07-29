FROM golang:1.9.7-alpine3.6

MAINTAINER "supnobita@gmail.com"

ADD server /
RUN mkdir data
VOLUME [ "data" ]
CMD ["/server"]

EXPOSE 8080