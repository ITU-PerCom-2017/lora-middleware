package main

import (
	"bytes"
	"net/http"
)

func DoPost(postUrl string, postbody string) (int, error) {

	url := postUrl
	//fmt.Println("URL:>", url)
	//fmt.Printf("POSTBODY: %s \n", postbody)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postbody)))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 404, err
	}
	defer resp.Body.Close()

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	//body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("response Body:", string(body))
	return resp.StatusCode, err
}
