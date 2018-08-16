FROM alpine:3.7

MAINTAINER "supnobita@gmail.com"

ADD server /
RUN mkdir data
VOLUME [ "/data" ]
CMD ["/server"]

EXPOSE 8080
