package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func browserlessRequest(
	browserlessToken string,
	url string,
	endpoint string,
	preprocessor string,
) ([]byte, error) {
	payload := make(map[string]interface{})

	switch endpoint {
	case "content":
		payload["url"] = url
		break
	case "function":
		payload["context"] = map[string]string{"url": url}
		payload["code"] = minifyJavascript(
			fmt.Sprintf(
				"module.exports=async({page,context})=>{const{url}=context;await page.goto(url);%sconst data=await page.content();return{data,type:'application/html'}}",
				preprocessor,
			),
		)
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	log.Println(string(reqBody))

	resp, err := http.Post(
		fmt.Sprintf(
			"https://chrome.browserless.io/%s?token=%s&headless=true&blockAds=true&stealth=true",
			endpoint,
			browserlessToken,
		),
		"application/json",
		bytes.NewBuffer(reqBody),
	)

	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func GrabContent(browserlessToken string, url string) ([]byte, error) {
	return browserlessRequest(browserlessToken, url, "content", "")
}

func GrabContentPreprocessing(
	browserlessToken string,
	url string,
	jsCode string,
) ([]byte, error) {
	return browserlessRequest(browserlessToken, url, "function", jsCode)
}

func minifyJavascript(jsCode string) string {
	return jsCode
}
