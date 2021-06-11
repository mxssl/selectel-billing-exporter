# Selectel Billing Exporter

Prometheus exporter для получения информации по биллингу аккаунта облака [Selectel](https://selectel.ru).

## Как работает экспортер

Экспортер раз в час ходит по url `https://my.selectel.ru/api/v3/billing/balance` с токеном в запросе, получает в json формате инфу по балансу средств на счете и отдает ее по url `/metrics` в формате prometheus.

Для работы экспортера нужно получить API [токен](https://my.selectel.ru/profile/apikeys):

## Как запустить

Создаем `docker-compose.yml` файл:

```yaml
version: '3'

services:
  selectel_exporter:
    image: mxssl/selectel_billing_exporter:0.0.2
    ports:
      - "6789:80"
    restart: always
    environment:
      TOKEN: тут_указываем_токен
```

Запускаем экспортер:

```sh
docker-compose up -d
```

Проверить работу экспортера:

```sh
docker-compose ps
docker-compose logs
```

Метрики доступны по url `your_ip:6789/metrics`

## Настройка для prometheus

```yaml
  - job_name: 'selectel_billing'
    scrape_interval: 60m
    static_configs:
      - targets: ['exporter_ip:6789']
```

## Пример алерта для alertmanager

```yaml
- alert: selectel_billing
  expr: selectel_billing_vpc_main{job="selectel_billing"} / 100 < 30000
  for: 180s
  labels:
    severity: warning
  annotations:
    summary: "{{ $labels.instance }}: В облаке Selectel на счете VPC меньше 30 тыс рублей"
    description: "Необходимо пополнить счет облака Selectel"
```

## Дашборд для графаны

[Дашборд](https://grafana.com/dashboards/9315)
