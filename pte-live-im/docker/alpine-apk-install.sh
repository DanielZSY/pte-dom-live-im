#!/bin/sh
# Alpine apk 安装（多镜像回退，Dockerfile 构建阶段调用）
# 用法: alpine-apk-install.sh [primary_mirror] pkg [pkg...]
# 示例: alpine-apk-install.sh mirrors.aliyun.com ca-certificates wget
# 海外: alpine-apk-install.sh dl-cdn.alpinelinux.org ca-certificates

set -eu

PRIMARY="${1:-mirrors.aliyun.com}"
shift

if [ "$#" -lt 1 ]; then
	echo "usage: alpine-apk-install.sh [mirror] pkg ..." >&2
	exit 1
fi

MIRROR_SCRIPT="/tmp/alpine-apk-mirror.sh"
if [ ! -x "${MIRROR_SCRIPT}" ]; then
	MIRROR_SCRIPT="$(dirname "$0")/alpine-apk-mirror.sh"
fi

if [ ! -f /etc/apk/repositories.orig ]; then
	cp /etc/apk/repositories /etc/apk/repositories.orig
fi

try_mirror() {
	m="$1"
	shift
	cp /etc/apk/repositories.orig /etc/apk/repositories
	"${MIRROR_SCRIPT}" "${m}"
	apk update
	apk add --no-cache "$@"
}

# 阿里云 aarch64 偶发 404/timeout；依次尝试国内镜像与官方源
for m in "${PRIMARY}" mirrors.tencent.com dl-cdn.alpinelinux.org; do
	echo ">> apk install try mirror=${m} packages=$*"
	if try_mirror "${m}" "$@"; then
		echo ">> apk install ok (mirror=${m})"
		exit 0
	fi
	echo ">> apk install failed on ${m}, fallback..." >&2
done

echo "error: apk install failed after all mirrors" >&2
exit 1
