FROM golang:alpine

RUN apk --no-cache add make git curl bash fish

WORKDIR /project

COPY kmip-go /project
RUN make tools

CMD make
