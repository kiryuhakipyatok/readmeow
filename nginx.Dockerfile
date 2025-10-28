FROM nginx:1.29.2-alpine

RUN rm /etc/nginx/conf.d/default.conf

COPY configs/nginx/default.conf /etc/nginx/conf.d/default.conf

RUN mkdir -p /etc/letsencrypt/live/api.readmeow.tech/
# RUN mkdir -p /var/www/certbot