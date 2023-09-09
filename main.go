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
	Status string `json:"status"`
	Data   struct {
		Currency  string `json:"currency"`
		IsPostpay bool   `json:"is_postpay"`
		Discount  int    `json:"discount"`
		Primary   struct {
			Main  int `json:"main"`
			Bonus int `json:"bonus"`
			VkRub int `json:"vk_rub"`
			Ref   int `json:"ref"`
			Hold  struct {
				Main  int `json:"main"`
				Bonus int `json:"bonus"`
				VkRub int `json:"vk_rub"`
			} `json:"hold"`
		} `json:"primary"`
		Storage struct {
			Main       int         `json:"main"`
			Bonus      int         `json:"bonus"`
			VkRub      int         `json:"vk_rub"`
			Prediction interface{} `json:"prediction"`
			Debt       int         `json:"debt"`
			Sum        int         `json:"sum"`
		} `json:"storage"`
		Vpc struct {
			Main       int         `json:"main"`
			Bonus      int         `json:"bonus"`
			VkRub      int         `json:"vk_rub"`
			Prediction interface{} `json:"prediction"`
			Debt       int         `json:"debt"`
			Sum        int         `json:"sum"`
		} `json:"vpc"`
		Vmware struct {
			Main       int         `json:"main"`
			Bonus      int         `json:"bonus"`
			VkRub      int         `json:"vk_rub"`
			Prediction interface{} `json:"prediction"`
			Debt       int         `json:"debt"`
			Sum        int         `json:"sum"`
		} `json:"vmware"`
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
	selectelNames := make(map[string][]string)
	selectelStructFields := []string{"main", "bonus", "vk_rub", "debt", "sum"}
	selectelNames["primary"] = []string{"main", "bonus", "vk_rub", "ref", "hold_main", "hold_bonus", "hold_vk_rub"}
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
		s := selectelBillingResponse{}

		// делаем запрос к Selectel API
		if err := getSelectelBilling(&s); err != nil {
			log.Printf("Не удалось получить ответ от Selectel API! Ошибка: %v\n", err)
			continue
		}

		// primary
		gauge["selectel_billing_primary_main"].Set(float64(s.Data.Primary.Main))
		gauge["selectel_billing_primary_bonus"].Set(float64(s.Data.Primary.Bonus))
		gauge["selectel_billing_primary_vk_rub"].Set(float64(s.Data.Primary.VkRub))
		gauge["selectel_billing_primary_ref"].Set(float64(s.Data.Primary.Ref))
		gauge["selectel_billing_primary_hold_main"].Set(float64(s.Data.Primary.Hold.Main))
		gauge["selectel_billing_primary_hold_bonus"].Set(float64(s.Data.Primary.Hold.Bonus))
		gauge["selectel_billing_primary_hold_vk_rub"].Set(float64(s.Data.Primary.Hold.VkRub))

		// storage
		gauge["selectel_billing_storage_main"].Set(float64(s.Data.Storage.Main))
		gauge["selectel_billing_storage_bonus"].Set(float64(s.Data.Storage.Bonus))
		gauge["selectel_billing_storage_vk_rub"].Set(float64(s.Data.Storage.VkRub))
		gauge["selectel_billing_storage_debt"].Set(float64(s.Data.Storage.Debt))
		gauge["selectel_billing_storage_sum"].Set(float64(s.Data.Storage.Sum))

		// vpc
		gauge["selectel_billing_vpc_main"].Set(float64(s.Data.Vpc.Main))
		gauge["selectel_billing_vpc_bonus"].Set(float64(s.Data.Vpc.Bonus))
		gauge["selectel_billing_vpc_vk_rub"].Set(float64(s.Data.Vpc.VkRub))
		gauge["selectel_billing_vpc_debt"].Set(float64(s.Data.Vpc.Debt))
		gauge["selectel_billing_vpc_sum"].Set(float64(s.Data.Vpc.Sum))

		// vmware
		gauge["selectel_billing_vmware_main"].Set(float64(s.Data.Vmware.Main))
		gauge["selectel_billing_vmware_bonus"].Set(float64(s.Data.Vmware.Bonus))
		gauge["selectel_billing_vmware_vk_rub"].Set(float64(s.Data.Vmware.VkRub))
		gauge["selectel_billing_vmware_debt"].Set(float64(s.Data.Vmware.Debt))
		gauge["selectel_billing_vmware_sum"].Set(float64(s.Data.Vmware.Sum))

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
