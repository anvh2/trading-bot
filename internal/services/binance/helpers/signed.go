package helpers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type SignedData struct {
	Body    *bytes.Buffer
	Header  http.Header
	FullURL string
}

func Signed(method string, fullURL string, params *url.Values) (*SignedData, error) {
	var (
		bodyStr  string = ""
		queryStr string = ""
	)

	if params != nil {
		params.Set("timestamp", fmt.Sprint(time.Now().UnixMilli()))
	}

	header := http.Header{}
	header.Set("X-MBX-APIKEY", os.Getenv("ORDER_API_KEY"))

	if params != nil {
		switch method {
		case http.MethodGet:
			queryStr = params.Encode()

		case http.MethodPost:
			bodyStr = params.Encode()
			header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	mac := hmac.New(sha256.New, []byte(os.Getenv("ORDER_SECRET_KEY")))
	_, err := mac.Write([]byte(fmt.Sprintf("%s%s", queryStr, bodyStr)))
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("signature", fmt.Sprintf("%x", mac.Sum(nil)))

	if queryStr == "" {
		queryStr = v.Encode()
	} else {
		queryStr = fmt.Sprintf("%s&%s", queryStr, v.Encode())
	}

	if queryStr != "" {
		fullURL = fmt.Sprintf("%s?%s", fullURL, queryStr)
	}

	return &SignedData{
		Body:    bytes.NewBufferString(bodyStr),
		Header:  header,
		FullURL: fullURL,
	}, nil
}
