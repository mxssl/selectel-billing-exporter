# Selectel Billing Exporter

Prometheus exporter для получения информации по биллингу аккаунта облака [Selectel](https://selectel.ru).

## Как работает экспортер

Экспортер раз в час ходит по url `https://api.selectel.ru/v3/balances` с токеном в запросе, получает в json формате инфу по балансу средств на счете и отдает ее по url `/metrics` в формате prometheus.

Для работы экспортера нужно получить API [токен](https://my.selectel.ru/profile/apikeys)

## Как запустить

### Запуск с помощью docker-compose

Создаем `docker-compose.yml` файл:

```yaml
version: '3'
services:
  selectel_exporter:
    image: mxssl/selectel-billing-exporter:1.1.4
    ports:
      - "6789:80"
    restart: always
    environment:
      TOKEN: <тут_указываем_токен>
```

Запускаем экспортер:

```sh
docker compose up -d
```

Проверить работу экспортера:

```sh
docker compose ps
docker compose logs
```

Метрики доступны по url `your_ip:6789/metrics`

## Kubernetes

### helm

[Установка helm чарта](https://github.com/mxssl/helm-charts/tree/main/charts/selectel-billing-exporter)

### Создание манифестов вручную

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: selectel-billing
  namespace: exporters
spec:
  selector:
    matchLabels:
      component: selectel-billing
  template:
    metadata:
      labels:
        component: selectel-billing
    spec:
      containers:
        - name: exporter
          image: mxssl/selectel-billing-exporter:1.1.4
          command: ["./app"]
          ports:
            - containerPort: 80
          env:
            - name: TOKEN
              value: <YOUR-TOKEN>

---
apiVersion: v1
kind: Service
metadata:
  name: selectel-billing
  namespace: exporters
spec:
  ports:
    - name: exporter
      port: 6789
      targetPort: 80
  selector:
    component: selectel-billing
```

```sh
kubectl apply -n exporters -f your-file.yaml
```

Внутри кластера метрики будут доступны по адресу `selectel-billing.exporters.svc.cluster.local:6789/metrics`

## Настройка для prometheus

```yaml
  - job_name: 'selectel_billing'
    scrape_interval: 60m
    static_configs:
      - targets: ['exporter_address:6789']
```

## Пример алерта для alertmanager

```yaml
- alert: selectel_billing
  expr: selectel_billing_final_sum{job="selectel_billing"} / 100 < 30000
  for: 180s
  labels:
    severity: warning
  annotations:
    summary: "{{ $labels.instance }}: В хостинге Selectel на счете меньше 30 000 рублей"
    description: "Необходимо пополнить счет в хостинге Selectel"
```

## Дашборд для графаны

[Дашборд](https://grafana.com/dashboards/9315)
