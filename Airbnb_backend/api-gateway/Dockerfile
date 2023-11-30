FROM nginx:latest

COPY ./api_gateway.conf /etc/nginx/conf.d/api_gateway.conf
COPY api-gateway.crt /etc/ssl/certs/api-gateway.crt
COPY api-gateway.key /etc/ssl/private/api-gateway.key
RUN rm /etc/nginx/conf.d/default.conf
