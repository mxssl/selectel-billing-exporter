# Selectel Billing Exporter

Прометеус экспортер для получения кол-ва средств на счете аккаунта Selectel.

## Как работает экспортер

Экспортер раз в час ходит по url `https://my.selectel.ru/api/v2/billing/balance` с токеном в запросе, получает в json формате инфу по балансу средств на счете и отдает ее по url `/metrics` в формате прометеуса.

Для работы экспортера нужно получить API [токен](https://kb.selectel.ru/24381209.html):

> Прежде чем приступать к работе с API, необходимо получить ключ (токен). Зарегистрированные пользователи Selectel могут получить ключ на странице my.selectel.ru/profile/apikeys. Токен представляет собой строку вида qX3Npu42ua73kPkhe4QCQ8Vv9_xxxxx, где xxxxx — это номер учётной записи пользователя.

## Как запустить

Создаем `docker-compose.yml` файл:

```
version: '3'

services:
  exporter:
    build: .
    image: mxssl/selectel_billing_exporter
    ports:
      - "6789:80"
    restart: always
    environment:
      TOKEN: тут_указываем_токен
```

Далее запускаем экспортер:

```
docker-compose up -d
```

Проверить работу экспортера можно следующими командами:

```
docker-compose ps
docker-compose logs
```

Метрики доступны по url `your_ip:6789/metrics`

## Настройка для prometheus:

```
  - job_name: 'selectel_billing'
    scrape_interval: 60m
    static_configs:
      - targets: ['exporter_ip:6789']
```

## Дашборд для графаны:

https://
