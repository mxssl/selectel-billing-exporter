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

	httpmock.RegisterResponder("GET", "https://my.selectel.ru/api/v3/billing/balance",
		httpmock.NewStringResponder(200, string(testdata)))

	expected := selectelBillingResponse{Status: "success", Data: struct {
		Currency  string "json:\"currency\""
		IsPostpay bool   "json:\"is_postpay\""
		Discount  int    "json:\"discount\""
		Primary   struct {
			Main  int "json:\"main\""
			Bonus int "json:\"bonus\""
			VkRub int "json:\"vk_rub\""
			Ref   int "json:\"ref\""
			Hold  struct {
				Main  int "json:\"main\""
				Bonus int "json:\"bonus\""
				VkRub int "json:\"vk_rub\""
			} "json:\"hold\""
		} "json:\"primary\""
		Storage struct {
			Main       int         "json:\"main\""
			Bonus      int         "json:\"bonus\""
			VkRub      int         "json:\"vk_rub\""
			Prediction interface{} "json:\"prediction\""
			Debt       int         "json:\"debt\""
			Sum        int         "json:\"sum\""
		} "json:\"storage\""
		Vpc struct {
			Main       int         "json:\"main\""
			Bonus      int         "json:\"bonus\""
			VkRub      int         "json:\"vk_rub\""
			Prediction interface{} "json:\"prediction\""
			Debt       int         "json:\"debt\""
			Sum        int         "json:\"sum\""
		} "json:\"vpc\""
		Vmware struct {
			Main       int         "json:\"main\""
			Bonus      int         "json:\"bonus\""
			VkRub      int         "json:\"vk_rub\""
			Prediction interface{} "json:\"prediction\""
			Debt       int         "json:\"debt\""
			Sum        int         "json:\"sum\""
		} "json:\"vmware\""
	}{Currency: "rub", IsPostpay: false, Discount: 0, Primary: struct {
		Main  int "json:\"main\""
		Bonus int "json:\"bonus\""
		VkRub int "json:\"vk_rub\""
		Ref   int "json:\"ref\""
		Hold  struct {
			Main  int "json:\"main\""
			Bonus int "json:\"bonus\""
			VkRub int "json:\"vk_rub\""
		} "json:\"hold\""
	}{Main: 10000, Bonus: 10000, VkRub: 0, Ref: 0, Hold: struct {
		Main  int "json:\"main\""
		Bonus int "json:\"bonus\""
		VkRub int "json:\"vk_rub\""
	}{Main: 0, Bonus: 0, VkRub: 0}}, Storage: struct {
		Main       int         "json:\"main\""
		Bonus      int         "json:\"bonus\""
		VkRub      int         "json:\"vk_rub\""
		Prediction interface{} "json:\"prediction\""
		Debt       int         "json:\"debt\""
		Sum        int         "json:\"sum\""
	}{Main: 203005, Bonus: 0, VkRub: 0, Prediction: interface{}(nil), Debt: 0, Sum: 203005}, Vpc: struct {
		Main       int         "json:\"main\""
		Bonus      int         "json:\"bonus\""
		VkRub      int         "json:\"vk_rub\""
		Prediction interface{} "json:\"prediction\""
		Debt       int         "json:\"debt\""
		Sum        int         "json:\"sum\""
	}{Main: 11250838, Bonus: 12345, VkRub: 0, Prediction: interface{}(nil), Debt: 0, Sum: 11150838}, Vmware: struct {
		Main       int         "json:\"main\""
		Bonus      int         "json:\"bonus\""
		VkRub      int         "json:\"vk_rub\""
		Prediction interface{} "json:\"prediction\""
		Debt       int         "json:\"debt\""
		Sum        int         "json:\"sum\""
	}{Main: 10000, Bonus: 10000, VkRub: 0, Prediction: interface{}(nil), Debt: 0, Sum: 20000}}}

	actual := selectelBillingResponse{}

	if err := getSelectelBilling(&actual); err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, actual)
}
