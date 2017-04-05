package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func DoPost(a *appContext, postUrl string, postbody string) (int, error) {

	url := postUrl
	//fmt.Println("URL:>", url)
	//fmt.Printf("POSTBODY: %s \n", postbody)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postbody)))
	req.SetBasicAuth(*a.Config.Giles.GilesUsername, *a.Config.Giles.GilesPassword)
	req.Header.Set("Content-Type", "text")
	req.Header.Set("Content-Length", fmt.Sprintf("%v", len(postbody)))
	//fmt.Println(*config.Giles.GilesPassword)
	//fmt.Println(*config.Giles.GilesUsername)
	resp, err := a.Client.Do(req)
	if err != nil {
		return 404, err
	}
	defer resp.Body.Close()

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	ioutil.ReadAll(resp.Body)
	//fmt.Println("response Body:", string(body))
	//resp.Body.Close()
	return resp.StatusCode, err
}
