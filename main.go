package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
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

type selectelBillingResponce struct {
	Status string `json:"status"`
	Data   struct {
		Currency string `json:"currency"`
		Selectel struct {
			Bonus   int `json:"bonus"`
			Voice   int `json:"voice"`
			Balance int `json:"balance"`
		} `json:"selectel"`
		Storage struct {
			Bonus   int `json:"bonus"`
			Voice   int `json:"voice"`
			Sum     int `json:"sum"`
			Balance int `json:"balance"`
			Debt    int `json:"debt"`
		} `json:"storage"`
		Vmware struct {
			Bonus   int `json:"bonus"`
			Voice   int `json:"voice"`
			Sum     int `json:"sum"`
			Balance int `json:"balance"`
			Debt    int `json:"debt"`
		} `json:"vmware"`
		Vpc struct {
			Bonus   int `json:"bonus"`
			Voice   int `json:"voice"`
			Sum     int `json:"sum"`
			Balance int `json:"balance"`
			Debt    int `json:"debt"`
		} `json:"vpc"`
	} `json:"data"`
}

func main() {
	log.Println("Selectel Billing Exporter запущен")

	log.Println("Устанавливаем Selectel токен...")
	ok := false
	TOKEN, ok = os.LookupEnv("TOKEN")
	if ok != true {
		log.Fatal("Please set environment variable TOKEN that contains Selectel API token")
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
	selectelNames := make(map[string][]string)
	selectelStructFields := []string{"bonus", "voice", "sum", "balance", "dept"}
	selectelNames["selectel"] = []string{"bonus", "voice", "balance"}
	selectelNames["storage"] = selectelStructFields
	selectelNames["vmware"] = selectelStructFields
	selectelNames["vpc"] = selectelStructFields

	promGauges := make(map[string]prometheus.Gauge)

	for name, fields := range selectelNames {
		for _, field := range fields {
			promGauges["selectel_billing_"+name+"_"+field] = prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "selectel_billing_" + name + "_" + field,
				Help: "selectel billing " + name + " " + field,
			})
		}
	}

	for _, v := range promGauges {
		prometheus.MustRegister(v)
	}
	return promGauges
}

func recordMetrics() {
	gauge := initGauges()
	for {
		s := selectelBillingResponce{}

		// делаем запрос к Selectel API
		if err := getSelectelBilling(&s); err != nil {
			log.Printf("Не удалось получить ответ от Selectel API! Ошибка: %v\n", err)
			continue
		}

		// selectel
		gauge["selectel_billing_selectel_bonus"].Set(float64(s.Data.Selectel.Bonus))
		gauge["selectel_billing_selectel_voice"].Set(float64(s.Data.Selectel.Voice))
		gauge["selectel_billing_selectel_balance"].Set(float64(s.Data.Selectel.Balance))

		// VPC
		gauge["selectel_billing_vpc_bonus"].Set(float64(s.Data.Vpc.Bonus))
		gauge["selectel_billing_vpc_voice"].Set(float64(s.Data.Vpc.Voice))
		gauge["selectel_billing_vpc_sum"].Set(float64(s.Data.Vpc.Sum))
		gauge["selectel_billing_vpc_balance"].Set(float64(s.Data.Vpc.Balance))
		gauge["selectel_billing_vpc_debt"].Set(float64(s.Data.Vpc.Debt))

		// storage
		gauge["selectel_billing_storage_bonus"].Set(float64(s.Data.Storage.Bonus))
		gauge["selectel_billing_storage_voice"].Set(float64(s.Data.Storage.Voice))
		gauge["selectel_billing_storage_sum"].Set(float64(s.Data.Storage.Sum))
		gauge["selectel_billing_storage_balance"].Set(float64(s.Data.Storage.Balance))
		gauge["selectel_billing_storage_debt"].Set(float64(s.Data.Storage.Debt))

		// vmware
		gauge["selectel_billing_vmware_bonus"].Set(float64(s.Data.Vmware.Bonus))
		gauge["selectel_billing_vmware_voice"].Set(float64(s.Data.Vmware.Voice))
		gauge["selectel_billing_vmware_sum"].Set(float64(s.Data.Vmware.Sum))
		gauge["selectel_billing_vmware_balance"].Set(float64(s.Data.Vmware.Balance))
		gauge["selectel_billing_vmware_debt"].Set(float64(s.Data.Vmware.Debt))

		time.Sleep(time.Hour * 1)
	}
}

func getSelectelBilling(selectelMetrics *selectelBillingResponce) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://my.selectel.ru/api/v2/billing/balance", nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-Token", TOKEN)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	temp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(temp, &selectelMetrics); err != nil {
		return err
	}

	return nil
}
