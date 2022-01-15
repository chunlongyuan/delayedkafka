FROM golang:1.13.7-alpine3.11 as builder
ARG COMPILER='go build'
COPY / /app
WORKDIR /app
ENV GO111MODULE=on GOPROXY="https://goproxy.cn,direct" CGO_ENABLED=0 GOSUMDB=off GOOS=linux

RUN echo "http://mirrors.aliyun.com/alpine/v3.11/main" > /etc/apk/repositories
RUN echo "http://mirrors.aliyun.com/alpine/v3.11/community" >> /etc/apk/repositories
RUN apk add -U tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime
RUN apk add --update --no-cache curl
RUN apk add bash build-base git
RUN git config --system http.sslverify false
RUN $COMPILER -o /go/bin/kdqueue .


FROM alpine:3.11
RUN echo "http://mirrors.aliyun.com/alpine/v3.11/main" > /etc/apk/repositories
RUN echo "http://mirrors.aliyun.com/alpine/v3.11/community" >> /etc/apk/repositories
RUN apk add -U tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime
RUN apk add --update --no-cache curl
COPY --from=builder /go/bin/kdqueue .