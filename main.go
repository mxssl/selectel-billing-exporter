package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TOKEN for selectel API
var TOKEN string

type selectelBillingResponse struct {
	Data struct {
		Billings []struct {
			FinalSum int `json:"final_sum"`
			DebtSum  int `json:"debt_sum"`
		} `json:"billings"`
	} `json:"data"`
}

func main() {
	log.Println("Selectel Billing Exporter запущен")

	log.Println("Устанавливаем Selectel токен...")
	ok := false
	TOKEN, ok = os.LookupEnv("TOKEN")
	if !ok {
		log.Fatal("Переменная окружения TOKEN, которая должна содержать Selectel API ключ не установлена")
	}
	log.Println("Токен успешно установлен!")

	http.Handle("/metrics", promhttp.Handler())

	go recordMetrics()

	srv := &http.Server{
		Addr: "0.0.0.0:80",

		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,

		Handler: nil,
	}

	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	log.Println("Экспортер готов принимать запросы от прометеуса на /metrics")

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	<-c

	log.Println("Изящно завершаем работу экспортера...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func initGauges() map[string]prometheus.Gauge {
	promGauges := make(map[string]prometheus.Gauge)

	promGauges["selectel_billing_final_sum"] = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_final_sum",
		Help: "selectel billing final sum",
	})

	promGauges["selectel_billing_debt_sum"] = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_debt_sum",
		Help: "selectel billing debt sum",
	})

	for _, v := range promGauges {
		prometheus.MustRegister(v)
	}
	return promGauges
}

func recordMetrics() {
	gauge := initGauges()
	for {
		s := selectelBillingResponse{}

		// делаем запрос к Selectel API
		if err := getSelectelBilling(&s); err != nil {
			log.Printf("Не удалось получить ответ от Selectel API! Ошибка: %v\n", err)
			continue
		}

		// записываем метрики
		gauge["selectel_billing_final_sum"].Set(float64(s.Data.Billings[0].FinalSum))
		gauge["selectel_billing_debt_sum"].Set(float64(s.Data.Billings[0].DebtSum))

		time.Sleep(time.Hour * 1)
	}
}

func getSelectelBilling(selectelMetrics *selectelBillingResponse) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.selectel.ru/v3/balances", nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-Token", TOKEN)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	temp, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(temp, &selectelMetrics); err != nil {
		return err
	}

	return nil
}
