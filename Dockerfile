FROM golang:1.13.7-alpine3.11 AS build
ENV GOPROXY=https://goproxy.cn
  
RUN mkdir -p /go/src/github.com/zdnscloud/immense
COPY . /go/src/github.com/zdnscloud/immense

WORKDIR /go/src/github.com/zdnscloud/immense
RUN CGO_ENABLED=0 GOOS=linux go build cmd/operator.go

FROM alpine:3.10.0

LABEL maintainers="Zdns Authors"
LABEL description="K8S Storage Operator"
RUN apk update && apk add udev blkid file util-linux e2fsprogs lvm2 udev sgdisk device-mapper python py-pip python-dev ceph --no-cache
RUN pip install prettytable
COPY --from=build /go/src/github.com/zdnscloud/immense/operator /operator
COPY deploy/ceph-key.py /ceph-key.py
ENTRYPOINT ["/bin/sh"]
