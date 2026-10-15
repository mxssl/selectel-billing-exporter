package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TOKEN для Selectel API
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
	if !ok || strings.TrimSpace(TOKEN) == "" {
		log.Fatal("Переменная окружения TOKEN, которая должна содержать Selectel API ключ не установлена")
	}
	log.Println("Токен успешно установлен!")

	http.Handle("/metrics", promhttp.Handler())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	gauges := initGauges()
	go recordMetrics(ctx, gauges, time.Hour)

	srv := &http.Server{
		Addr: "0.0.0.0:80",

		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,

		Handler: nil,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка HTTP-сервера: %v", err)
		}
	}()

	log.Println("Экспортер готов принимать запросы от Prometheus на /metrics")

	<-ctx.Done()

	log.Println("Изящно завершаем работу экспортера...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		log.Fatal(err)
	}
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

func recordMetrics(ctx context.Context, gauges map[string]prometheus.Gauge, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		var response selectelBillingResponse
		if err := getSelectelBilling(ctx, &response); err != nil {
			log.Printf("Не удалось получить ответ от Selectel API: %v", err)
		} else {
			gauges["selectel_billing_final_sum"].Set(float64(response.Data.Billings[0].FinalSum))
			gauges["selectel_billing_debt_sum"].Set(float64(response.Data.Billings[0].DebtSum))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func getSelectelBilling(ctx context.Context, selectelMetrics *selectelBillingResponse) error {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.selectel.ru/v3/balances", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Token", TOKEN)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("selectel API returned HTTP %d", resp.StatusCode)
	}
	var response selectelBillingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("decode Selectel billing response: %w", err)
	}
	if len(response.Data.Billings) == 0 {
		return errors.New("selectel API returned no billing accounts")
	}
	*selectelMetrics = response
	return nil
}
