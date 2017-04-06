package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var files []string

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

func GetMetadata(path string) map[string]Metadatum {
	var mm map[string]Metadatum
	mm = make(map[string]Metadatum)

	err := filepath.Walk(path, addFiles)
	if err != nil {
		log.Fatal(err)
	}

	for _, filename := range files {
		file, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Printf("File error: %v\n", err)
			os.Exit(1)
		}
		var m Metadatum
		if err = json.Unmarshal(file, &m); err != nil {
			fmt.Printf("Unmarshal error: %v: %v\n", err, filename)
		} else {
			fmt.Printf("%v\n", m)

			extra := m.Extra.(map[string]interface{})
			device_name := extra["Device_id"].(string)
			//device_name := string(m.Extra["Device_id"])
			mm[device_name] = m
		}
	}
	return mm

}

func addFiles(path string, f os.FileInfo, err error) error {
	if strings.Contains(path, "json") {
		files = append(files, path)
	}
	return nil
}
