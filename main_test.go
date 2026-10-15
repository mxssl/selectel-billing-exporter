package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestSelectelBillingRequest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	testdata, err := os.ReadFile("testdata/SelectelBillingRespExample.json")
	if err != nil {
		t.Fatal(err)
	}

	httpmock.RegisterResponder("GET", "https://api.selectel.ru/v3/balances",
		httpmock.NewStringResponder(200, string(testdata)))

	expected := selectelBillingResponse{
		Data: struct {
			Billings []struct {
				FinalSum int `json:"final_sum"`
				DebtSum  int `json:"debt_sum"`
			} `json:"billings"`
		}{
			Billings: []struct {
				FinalSum int `json:"final_sum"`
				DebtSum  int `json:"debt_sum"`
			}{
				{
					FinalSum: 33218108,
					DebtSum:  0,
				},
			},
		},
	}

	actual := selectelBillingResponse{}

	if err := getSelectelBilling(context.Background(), &actual); err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, actual)
}

func TestSelectelBillingErrors(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"unauthorized", 401, `{"error":"unauthorized"}`},
		{"server error", 503, `unavailable`},
		{"invalid JSON", 200, `{`},
		{"empty billing", 200, `{"data":{"billings":[]}}`},
		{"missing data", 200, `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			httpmock.RegisterResponder("GET", "https://api.selectel.ru/v3/balances", httpmock.NewStringResponder(tc.status, tc.body))
			var response selectelBillingResponse
			assert.Error(t, getSelectelBilling(context.Background(), &response))
		})
	}
}

func TestRecordMetricsWaitsAfterFailure(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	called := make(chan struct{}, 10)
	httpmock.RegisterResponder("GET", "https://api.selectel.ru/v3/balances", func(req *http.Request) (*http.Response, error) {
		called <- struct{}{}
		return httpmock.NewStringResponse(503, "unavailable"), nil
	})
	done := make(chan struct{})
	go func() { recordMetrics(ctx, nil, time.Hour); close(done) }()
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("initial request was not made")
	}
	select {
	case <-called:
		t.Error("retried without waiting")
	case <-time.After(50 * time.Millisecond):
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("poller did not stop")
	}
}

func TestSelectelBillingRequestCancellation(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	// Rewrite the fixed API endpoint to a local TLS server using a dedicated transport.
	original := http.DefaultTransport
	http.DefaultTransport = rewriteTransport{target: server.URL, transport: server.Client().Transport}
	defer func() { http.DefaultTransport = original }()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	var response selectelBillingResponse
	assert.ErrorIs(t, getSelectelBilling(ctx, &response), context.DeadlineExceeded)
}

type rewriteTransport struct {
	target    string
	transport http.RoundTripper
}

func (rt rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	target, err := url.Parse(rt.target)
	if err != nil {
		return nil, err
	}
	req = req.Clone(req.Context())
	req.URL.Scheme, req.URL.Host = target.Scheme, target.Host
	return rt.transport.RoundTrip(req)
}

func TestRecordMetricsPreservesBalanceAfterFailure(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gauges := map[string]prometheus.Gauge{
		"selectel_billing_final_sum": prometheus.NewGauge(prometheus.GaugeOpts{Name: "selectel_billing_final_sum"}),
		"selectel_billing_debt_sum":  prometheus.NewGauge(prometheus.GaugeOpts{Name: "selectel_billing_debt_sum"}),
	}
	calls := 0
	httpmock.RegisterResponder("GET", "https://api.selectel.ru/v3/balances", func(req *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return httpmock.NewStringResponse(200, `{"data":{"billings":[{"final_sum":12345,"debt_sum":67}]}}`), nil
		}
		cancel()
		return httpmock.NewStringResponse(503, "unavailable"), nil
	})
	done := make(chan struct{})
	go func() { recordMetrics(ctx, gauges, 10*time.Millisecond); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("poller did not stop")
	}
	for name, expected := range map[string]float64{"selectel_billing_final_sum": 12345, "selectel_billing_debt_sum": 67} {
		metric := &dto.Metric{}
		if err := gauges[name].Write(metric); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expected, metric.GetGauge().GetValue())
	}
}

func TestSelectelBillingClosesResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	for _, status := range []int{200, 503} {
		body := &trackedBody{Reader: strings.NewReader(`{"data":{"billings":[{"final_sum":1,"debt_sum":0}]}}`)}
		httpmock.RegisterResponder("GET", "https://api.selectel.ru/v3/balances", func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "test-token", req.Header.Get("X-Token"))
			response := httpmock.NewStringResponse(status, "")
			response.Body = body
			return response, nil
		})
		oldToken := TOKEN
		TOKEN = "test-token"
		var response selectelBillingResponse
		err := getSelectelBilling(context.Background(), &response)
		TOKEN = oldToken
		if status == 200 {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
		assert.True(t, body.closed)
	}
}

type trackedBody struct {
	*strings.Reader
	closed bool
}

func (body *trackedBody) Close() error { body.closed = true; return nil }
