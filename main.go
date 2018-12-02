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

func init() {
	prometheus.MustRegister(selectelBillingSelectelBonus)
	prometheus.MustRegister(selectelBillingSelectelVoice)
	prometheus.MustRegister(selectelBillingSelectelBalance)

	prometheus.MustRegister(selectelBillingVPCBonus)
	prometheus.MustRegister(selectelBillingVPCVoice)
	prometheus.MustRegister(selectelBillingVPCSum)
	prometheus.MustRegister(selectelBillingVPCBalance)
	prometheus.MustRegister(selectelBillingVPCDebt)

	prometheus.MustRegister(selectelBillingStorageBonus)
	prometheus.MustRegister(selectelBillingStorageVoice)
	prometheus.MustRegister(selectelBillingStorageSum)
	prometheus.MustRegister(selectelBillingStorageBalance)
	prometheus.MustRegister(selectelBillingStorageDebt)

	prometheus.MustRegister(selectelBillingVmwareBonus)
	prometheus.MustRegister(selectelBillingVmwareVoice)
	prometheus.MustRegister(selectelBillingVmwareSum)
	prometheus.MustRegister(selectelBillingVmwareBalance)
	prometheus.MustRegister(selectelBillingVmwareDebt)
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

func recordMetrics() {
	go func() {
		for {
			s := selectelBillingResponce{}

			// делаем запрос к Selectel API
			if err := getSelectelBilling(&s); err != nil {
				log.Printf("Не удалось получить ответ от Selectel API! Ошибка: %v\n", err)
			}

			// selectel
			selectelBillingSelectelBonus.Set(float64(s.Data.Selectel.Bonus))
			selectelBillingSelectelVoice.Set(float64(s.Data.Selectel.Voice))
			selectelBillingSelectelBalance.Set(float64(s.Data.Selectel.Balance))

			// VPC
			selectelBillingVPCBonus.Set(float64(s.Data.Vpc.Bonus))
			selectelBillingVPCVoice.Set(float64(s.Data.Vpc.Voice))
			selectelBillingVPCSum.Set(float64(s.Data.Vpc.Sum))
			selectelBillingVPCBalance.Set(float64(s.Data.Vpc.Balance))
			selectelBillingVPCDebt.Set(float64(s.Data.Vpc.Debt))

			// storage
			selectelBillingStorageBonus.Set(float64(s.Data.Storage.Bonus))
			selectelBillingStorageVoice.Set(float64(s.Data.Storage.Voice))
			selectelBillingStorageSum.Set(float64(s.Data.Storage.Sum))
			selectelBillingStorageBalance.Set(float64(s.Data.Storage.Balance))
			selectelBillingStorageDebt.Set(float64(s.Data.Storage.Debt))

			// vmware
			selectelBillingVmwareBonus.Set(float64(s.Data.Vmware.Bonus))
			selectelBillingVmwareVoice.Set(float64(s.Data.Vmware.Voice))
			selectelBillingVmwareSum.Set(float64(s.Data.Vmware.Sum))
			selectelBillingVmwareBalance.Set(float64(s.Data.Vmware.Balance))
			selectelBillingVmwareDebt.Set(float64(s.Data.Vmware.Debt))

			time.Sleep(time.Hour * 1)
		}
	}()
}

var (
	selectelBillingVPCBonus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_vpc_bonus",
		Help: "Selectel billing vpc bonus",
	})
)

var (
	selectelBillingVPCVoice = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_vpc_voice",
		Help: "Selectel billing vpc voice",
	})
)

var (
	selectelBillingVPCSum = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_vpc_sum",
		Help: "Selectel billing vpc sum",
	})
)

var (
	selectelBillingVPCBalance = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_vpc_balance",
		Help: "Selectel billing vpc balance",
	})
)

var (
	selectelBillingVPCDebt = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_vpc_debt",
		Help: "Selectel billing vpc debt",
	})
)

var (
	selectelBillingSelectelBonus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_bonus",
		Help: "Selectel billing bonus",
	})
)

var (
	selectelBillingSelectelVoice = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_voice",
		Help: "Selectel billing debt",
	})
)

var (
	selectelBillingSelectelBalance = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_balance",
		Help: "Selectel billing balance",
	})
)

var (
	selectelBillingStorageBonus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_storage_bonus",
		Help: "Selectel billing storage bonus",
	})
)

var (
	selectelBillingStorageVoice = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_storage_voice",
		Help: "Selectel billing storage voice",
	})
)

var (
	selectelBillingStorageSum = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_storage_sum",
		Help: "Selectel billing storage sum",
	})
)

var (
	selectelBillingStorageBalance = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_storage_balance",
		Help: "Selectel billing storage balance",
	})
)

var (
	selectelBillingStorageDebt = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_storage_debt",
		Help: "Selectel billing storage debt",
	})
)

var (
	selectelBillingVmwareBonus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_vmware_bonus",
		Help: "Selectel billing vmware bonus",
	})
)

var (
	selectelBillingVmwareVoice = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_vmware_voice",
		Help: "Selectel billing vmware voice",
	})
)

var (
	selectelBillingVmwareSum = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_vmware_sum",
		Help: "Selectel billing vmware sum",
	})
)

var (
	selectelBillingVmwareBalance = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_vmware_balance",
		Help: "Selectel billing vmware balance",
	})
)

var (
	selectelBillingVmwareDebt = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "selectel_billing_vmware_debt",
		Help: "Selectel billing vmware debt",
	})
)

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
