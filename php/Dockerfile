FROM php:8.2-apache

RUN sed -i 's/Listen 80/Listen 8000/' /etc/apache2/ports.conf
RUN sed -i 's/:80>/:8000>/' /etc/apache2/sites-enabled/000-default.conf

COPY . /var/www/html/

CMD ["apache2-foreground"]
