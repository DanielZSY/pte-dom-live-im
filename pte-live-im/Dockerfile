# syntax=docker/dockerfile:1

FROM golang:1.26.3-alpine AS build-dist
ENV GOPROXY=https://goproxy.cn,direct
ARG APK_MIRROR=mirrors.aliyun.com

COPY docker/alpine-apk-mirror.sh docker/alpine-apk-install.sh /tmp/
RUN chmod +x /tmp/alpine-apk-mirror.sh /tmp/alpine-apk-install.sh \
	&& /tmp/alpine-apk-install.sh "${APK_MIRROR}" git ca-certificates \
	&& rm /tmp/alpine-apk-mirror.sh /tmp/alpine-apk-install.sh

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
	go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	mkdir -p /app/build/dist && go build -mod=readonly -o /app/build/dist/pte_live_im ./cmd/main

FROM alpine:3.20 AS prod
ARG APK_MIRROR=mirrors.aliyun.com

COPY docker/alpine-apk-mirror.sh docker/alpine-apk-install.sh /tmp/
RUN chmod +x /tmp/alpine-apk-mirror.sh /tmp/alpine-apk-install.sh \
	&& /tmp/alpine-apk-install.sh "${APK_MIRROR}" ca-certificates wget \
	&& rm /tmp/alpine-apk-mirror.sh /tmp/alpine-apk-install.sh

WORKDIR /data/pte_live_im
COPY --from=build-dist /app/build/dist/pte_live_im ./

EXPOSE 11510 11511 11512

CMD ["/data/pte_live_im/pte_live_im", "-c", "/data/pte_live_im/conf/app.yaml"]
