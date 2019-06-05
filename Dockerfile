FROM golang:1.12.5-alpine3.9 AS build
  
RUN mkdir -p /go/src/github.com/zdnscloud/immense
COPY . /go/src/github.com/zdnscloud/immense

WORKDIR /go/src/github.com/zdnscloud/immense
RUN CGO_ENABLED=0 GOOS=linux go build cmd/operator.go

FROM alpine:3.9.4

LABEL maintainers="Zdns Authors"
LABEL description="K8S Storage Operator"
RUN apk update && apk add udev blkid file util-linux e2fsprogs lvm2 udev sgdisk device-mapper
COPY --from=build /go/src/github.com/zdnscloud/immense/operator /operator
ENTRYPOINT ["/bin/sh"]
