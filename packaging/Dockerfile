FROM golang:1.15-alpine AS build

RUN apk add build-base

COPY . /go/src/github.com/grupawp/tensorflow-deploy
ENV GOFLAGS -mod=vendor
RUN make -C /go/src/github.com/grupawp/tensorflow-deploy install

FROM alpine:3

COPY --from=build /go/bin/tensorflow_deploy /
COPY --from=build /go/bin/bootstrap /

COPY packaging/launch-tfd.sh /
COPY packaging/config.yaml /usr/local/etc/config.yaml

WORKDIR /tfdeploy

ENTRYPOINT [ "/launch-tfd.sh" ]
