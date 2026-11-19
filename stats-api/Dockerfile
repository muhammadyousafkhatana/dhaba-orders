FROM composer:2 AS composer
WORKDIR /app
COPY composer.json .
RUN composer install --no-dev --no-interaction --prefer-dist

FROM php:8.3-apache
RUN apt-get update && apt-get install -y libpq-dev \
    && docker-php-ext-install pdo pdo_pgsql \
    && rm -rf /var/lib/apt/lists/*
COPY --from=composer /app/vendor ./vendor
COPY public/ /var/www/html/
EXPOSE 80
