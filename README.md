# Selectel Billing Exporter

Prometheus exporter для получения информации по биллингу аккаунта облака [Selectel](https://selectel.ru).

## Как работает экспортер

Экспортер раз в час ходит по url `https://api.selectel.ru/v3/balances` с токеном в запросе, получает в json формате инфу по балансу средств на счете и отдает ее по url `/metrics` в формате prometheus.

При ошибке API экспортер сохраняет последние полученные значения и повторяет запрос через час.

Для работы экспортера нужно получить API [токен](https://my.selectel.ru/profile/apikeys)

## Как запустить

Текущий релиз: [1.1.7](https://github.com/mxssl/selectel-billing-exporter/releases/tag/1.1.7).
Docker-образ и готовый бинарный файл предназначены для Linux amd64.

### Установка бинарного файла

```sh
curl -fLO https://github.com/mxssl/selectel-billing-exporter/releases/download/1.1.7/selectel-billing-exporter-linux-amd64.tar.gz
curl -fLO https://github.com/mxssl/selectel-billing-exporter/releases/download/1.1.7/checksums.txt
sha256sum --ignore-missing -c checksums.txt
tar -xzf selectel-billing-exporter-linux-amd64.tar.gz
sudo install -m 0755 selectel-billing-exporter /usr/local/bin/selectel-billing-exporter
```

Задайте непустой `TOKEN` и запустите экспортер с правом слушать порт 80,
например через systemd с `AmbientCapabilities=CAP_NET_BIND_SERVICE`.
Метрики доступны по адресу `http://localhost/metrics`.

### Запуск с помощью docker-compose

Создаем `compose.yaml` файл:

```yaml
services:
  selectel_exporter:
    image: mxssl/selectel-billing-exporter:1.1.7
    ports:
      - "6789:80"
    restart: always
    environment:
      TOKEN: ${TOKEN:?Укажите Selectel API токен}
```

Запускаем экспортер:

```sh
export TOKEN='<ваш_Selectel_API_токен>'
docker compose pull
docker compose up -d
```

Проверить работу экспортера:

```sh
docker compose ps
docker compose logs
```

Метрики доступны по url `your_ip:6789/metrics`

Для обновления существующей установки укажите тег `1.1.7` в `compose.yaml`
и повторите `docker compose pull && docker compose up -d`.

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
          image: mxssl/selectel-billing-exporter:1.1.7
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
kubectl create namespace exporters --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -n exporters -f your-file.yaml
kubectl rollout status -n exporters deployment/selectel-billing
```

Внутри кластера метрики будут доступны по адресу `selectel-billing.exporters.svc.cluster.local:6789/metrics`

## Настройка для prometheus

```yaml
- job_name: "selectel_billing"
  scrape_interval: 60m
  static_configs:
    - targets: ["exporter_address:6789"]
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
