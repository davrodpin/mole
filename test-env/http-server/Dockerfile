FROM alpine:3.6

RUN apk update && apk add nginx curl
RUN mkdir -p /run/nginx
RUN mkdir -p /data/www

COPY default.conf /etc/nginx/conf.d/
COPY index.html /data/www

CMD nginx -g 'daemon off;'
