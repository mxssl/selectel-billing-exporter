package main

import (
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestSelectelBillingRequest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	testdata, err := os.ReadFile("testdata/SelectelBillingRespExample.json")
	if err != nil {
		t.Error(err)
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

	if err := getSelectelBilling(&actual); err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, actual)
}
