package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

func main() {
	readConfig()
	token := postAuth()
	paymentID := postPayment(token)
	postConfirmation(token, paymentID)
	postCancellation(token, paymentID)
}

func readConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	path, _ := os.Getwd()
	viper.AddConfigPath(path)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s \\n", err))
	}
}

func postAuth() tokenStruct {
	userName := viper.GetString("Auth.Username")
	userPassword := viper.GetString("Auth.Password")
	url := viper.GetString("Auth.Url")
	client := resty.New()

	var resp tokenStruct

	r, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBasicAuth(userName, userPassword).
		SetBody("grant_type=client_credentials").
		SetResult(&resp).
		Post(url)

	if err != nil {
		fmt.Println("Auth fail!", err)
	}

	if r.StatusCode() != 200 {
		fmt.Println("Auth fail!", r.StatusCode(), "on", url)
	}

	fmt.Println("Auth Success!")
	return resp
}

func postPayment(token tokenStruct) string {
	url := viper.GetString("BaseUrl") + viper.GetString("Payment")
	client := resty.New()

	var response paymentResponseStruct

	r, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token.AccessToken).
		SetBody(paymentModel).
		SetResult(&response).
		Post(url)

	if err != nil {
		fmt.Println("Payment Fail", url)
		return ""
	}

	if r.StatusCode() >= 400 {
		fmt.Println("Payment Fail ", r.StatusCode(), "on", url)
		return ""
	}

	fmt.Println("Payment Success ", response.Payment.PaymentID)

	return response.Payment.PaymentID
}

func postConfirmation(token tokenStruct, paymentID string) {
	url := viper.GetString("BaseUrl") + strings.Replace(viper.GetString("Confirmation"), "{{PaymentId}}", paymentID, 1)
	client := resty.New()

	var response confirmationResponseStruct

	r, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token.AccessToken).
		SetResult(&response).
		Put(url)

	if err != nil {
		fmt.Println("Confirmation Fail", url)
		return
	}

	if r.StatusCode() >= 400 {
		fmt.Println("Confirmation Fail ", r.StatusCode(), "on", url, response.errors)
		return
	}

	fmt.Println("Confirmation success")
}

func postCancellation(token tokenStruct, paymentID string) {
	url := viper.GetString("BaseUrl") + strings.Replace(viper.GetString("Cancellation"), "{{PaymentId}}", paymentID, 1)
	client := resty.New()

	r, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token.AccessToken).
		SetBody(cancellationModel).
		Post(url)

	if err != nil {
		fmt.Println("Cancellation Fail", url)
		return
	}

	if r.StatusCode() >= 400 {
		fmt.Println("Cancellation Fail ", r.StatusCode(), "on", url)
		return
	}

	fmt.Println("Cancellation success")
}

type tokenStruct struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type paymentResponseStruct struct {
	Payment paymentStruct `json:"Payment"`
}

type paymentStruct struct {
	PaymentID string `json:"PaymentId"`
}

const paymentModel = `{
		"MerchantOrderId": "1587997030607",
		"Payment": {
		  "Type": "PhysicalCreditCard",
		  "SoftDescriptor": "Desafio GO 2",
		  "PaymentDateTime": "2020-01-08T11:00:00",
		  "Amount": 100,
		  "Installments": 1,
		  "Interest": "ByMerchant",
		  "Capture": true,
		  "ProductId": 1,
		  "CreditCard": {
			"CardNumber": "5432123454321234",
			"ExpirationDate": "12/2021",
			"SecurityCodeStatus": "Collected",
			"SecurityCode": "123",
			"BrandId": 1,
			"IssuerId": 401,
			"InputMode": "Typed",
			"AuthenticationMethod": "NoPassword",
			"TruncateCardNumberWhenPrinting": true
		  },
		  "PinPadInformation": {
			"PhysicalCharacteristics": "PinPadWithChipReaderWithoutSamAndContactless",
			"ReturnDataInfo": "00",
			"SerialNumber": "0820471929",
			"TerminalId": "42004558"
		  }
		}
	  }`

const cancellationModel = `json:{	
		"MerchantVoidId": "1587997297176",
		"MerchantVoidDate": "2020-04-27T14:21:37.176Z",
		"Card": {
		  "InputMode": "Typed",
		  "CardNumber": "5432123454321234"
		}
	  }`

type confirmationResponseStruct struct {
	errors []errorStruct
}

type errorStruct struct {
	ReturnMessage string `json:"Code"`
	ReturnCode    string `json:"Message"`
}
