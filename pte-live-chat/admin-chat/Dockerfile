# syntax=docker/dockerfile:1
# 静态产物：本地 make pack → 上传 dist/ 到本目录；镜像仅 nginx + COPY

FROM nginx:1.27-alpine

ARG APK_MIRROR=mirrors.aliyun.com
COPY docker/alpine-apk-mirror.sh docker/alpine-apk-install.sh /tmp/
RUN chmod +x /tmp/alpine-apk-mirror.sh /tmp/alpine-apk-install.sh \
	&& /tmp/alpine-apk-install.sh "${APK_MIRROR}" wget \
	&& rm /tmp/alpine-apk-mirror.sh /tmp/alpine-apk-install.sh

COPY docker/nginx.conf.template /etc/nginx/templates/default.conf.template
COPY docker/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

COPY dist /usr/share/nginx/html

EXPOSE 80

ENTRYPOINT ["/docker-entrypoint.sh"]
