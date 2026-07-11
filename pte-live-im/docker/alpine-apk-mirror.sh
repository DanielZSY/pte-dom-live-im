#!/bin/sh
# Alpine apk 国内镜像（Dockerfile 构建阶段调用）
# 用法: alpine-apk-mirror.sh [mirrors.aliyun.com]
# 海外构建: --build-arg APK_MIRROR=dl-cdn.alpinelinux.org

set -eu

MIRROR="${1:-mirrors.aliyun.com}"

if [ -z "${MIRROR}" ] || [ "${MIRROR}" = "dl-cdn.alpinelinux.org" ]; then
	exit 0
fi

if [ ! -f /etc/apk/repositories ]; then
	echo "warn: /etc/apk/repositories not found" >&2
	exit 0
fi

sed -i "s#https\\?://dl-cdn.alpinelinux.org/#https://${MIRROR}/#g" /etc/apk/repositories
sed -i "s#https\\?://dl-[0-9]\\+\\.alpinelinux.org/#https://${MIRROR}/#g" /etc/apk/repositories

echo ">> Alpine apk mirror: ${MIRROR}"
