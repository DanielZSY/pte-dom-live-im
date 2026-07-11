#!/bin/sh
set -eu

if [ -f /etc/nginx/templates/default.conf.template ]; then
	cp /etc/nginx/templates/default.conf.template /etc/nginx/conf.d/default.conf
fi

exec nginx -g 'daemon off;'
