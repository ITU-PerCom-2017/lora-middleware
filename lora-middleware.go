/*
 * Copyright (c) 2013 IBM Corp.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	//"log"

	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/satori/go.uuid"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

//func onMessageReceived(client MQTT.Client, message MQTT.Message) {
//go giles_pup(message)
//}

func onMessageReceived(a *appContext) MQTT.MessageHandler {
	return MQTT.MessageHandler(func(client MQTT.Client, message MQTT.Message) {
		go giles_pup(message, a)
	})
}

//var NS uuid.UUID
//var UUIDS map[string]bool

//var log *logging.Logger

var config *Config

type appContext struct {
	Config   *Config
	UUID_NS  *uuid.UUID
	UUIDS    map[string]bool
	Client   *http.Client
	Metadata map[string]Metadatum
}

const LORA string = "[LORA]    "

func init() {
	MQTT.DEBUG = log.New(os.Stdout, "DEBUG ", 0)
	MQTT.WARN = log.New(os.Stdout, "WARN ", 0)
	MQTT.CRITICAL = log.New(os.Stdout, "CRITICAL ", 0)
	MQTT.ERROR = log.New(os.Stdout, "ERROR ", 0)
}

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("signal received, exiting")
		os.Exit(0)
	}()

	hostname, _ := os.Hostname()
	configfile := flag.String("configfile", "", "configfile")
	qos := flag.Int("qos", 0, "The QoS to subscribe to messages at")
	clientid := flag.String("clientid", hostname+strconv.Itoa(time.Now().Second()), "A clientid for the connection")
	flag.Parse()

	a := &appContext{}
	// client is safe for concurrent use, faster to use just one client
	// https://github.com/golang/go/issues/4049
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	a.Client = &http.Client{Transport: tr}
	a.Config = LoadConfig(*configfile)
	a.UUIDS = make(map[string]bool)
	NS, err := uuid.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		fmt.Printf("Something gone wrong: %s", err)
		os.Exit(0)
	}
	a.UUID_NS = &NS

	a.Metadata = GetMetadata(*a.Config.Metadata.MetadataFolder)

	connOpts := &MQTT.ClientOptions{
		ClientID:             *clientid,
		CleanSession:         true,
		Username:             *a.Config.TTN.MQTTUsername,
		Password:             *a.Config.TTN.MQTTPassword,
		MaxReconnectInterval: 1 * time.Second,
		//KeepAlive:            60 * time.Second,
		TLSConfig: tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert},
	}
	connOpts.AddBroker(*a.Config.TTN.MQTTServer)
	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(*a.Config.TTN.MQTTTopic, byte(*qos), onMessageReceived(a)); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to %s\n", *a.Config.TTN.MQTTServer)
	}

	for {
		time.Sleep(100 * time.Millisecond)
	}
}

func giles_pup(message MQTT.Message, a *appContext) {
	var m UplinkMessage
	s := message.Payload()

	if err := json.Unmarshal(s, &m); err != nil {
		fmt.Printf("Error with %s\n", s)
	} else {
		if len(m.PayloadRaw) == 3 {
			//fmt.Printf("%v\n", m.PayloadRaw)
			//fmt.Printf("%q\n", m.PayloadRaw)
			var sm SmapMeta
			var property Property
			var metadatum Metadatum
			gm, exists := a.Metadata[m.DevID]
			if exists {
				metadatum = gm
			} else {
				var location Location
				metadatum.Location = location
			}
			var instrument Instrument

			var model string
			var sensorType string
			var units string
			var found bool
			var v float64
			// temperature
			if m.PayloadRaw[0] == 170 { // Chirp Temperature Sensor
				model = "Chirp"
				sensorType = "Temperature"
				units = "C"
				v = float64((int(m.PayloadRaw[1])<<8)+int(m.PayloadRaw[2])) / 10
				found = true
			} else if m.PayloadRaw[0] == 187 { // Chirp Humidity Sensor
				model = "Chirp"
				sensorType = "Humidity"
				units = "Moisture"
				v = float64((int(m.PayloadRaw[1]) << 8) + int(m.PayloadRaw[2]))
				found = true
			} else if m.PayloadRaw[0] == 204 { // Chirp Light Sensor
				model = "Chirp"
				sensorType = "Light"
				units = "Light"
				v = float64((int(m.PayloadRaw[1]) << 8) + int(m.PayloadRaw[2]))
				v = 100 * ((65535 - v) / 65535)
				found = true
			} else if m.PayloadRaw[0] == 1 { // Sharp Dust Sensor
				model = "GP2Y1010AU0F"
				sensorType = "Dust"
				units = "Particles"
				v = float64((int(m.PayloadRaw[1]) << 8) + int(m.PayloadRaw[2]))
				found = true
				fmt.Printf("DUST\n\n\n")
			} else if m.PayloadRaw[0] == 2 { // PIR sensor
				model = "PIR Sensor"
				sensorType = "PIR"
				units = "Motion"
				v = float64((int(m.PayloadRaw[1]) << 8) + int(m.PayloadRaw[2]))
				found = true
			}
			if found {
				u1 := uuid.NewV5(*a.UUID_NS, m.DevID+model+sensorType).String()

				tr := make([][]float64, 1)
				tr[0] = make([]float64, 2)
				tr[0][0] = float64(m.Metadata.Time.Unix())
				tr[0][1] = float64(v)

				path := "ITU/5/5A56/" + m.DevID + "/" + model + "/" + sensorType

				_, exists := a.UUIDS[u1]

				if exists {
					// just post new reading
					sm.UUID = u1
					sm.Readings = tr

				} else {
					metadatum.Instrument = instrument
					sm.Metadata = &metadatum
					sm.Properties = &property
					sm.Metadata.Instrument.Model = model
					sm.Properties.UnitofMeasure = units
					sm.Path = path
					sm.UUID = u1
					sm.Metadata.SourceName = "Instrumentation_Test"
					sm.Properties.Timezone = "Europe/Copenhagen"
					sm.Metadata.Location.Building = "IT University of Copenhagen"
					sm.Metadata.Location.City = "Copenhagen"
					sm.Metadata.System = "Lora"
					sm.Metadata.Instrument.Manufacturer = "LoPy"
					sm.Readings = tr
				}

				jsonStr, err := json.Marshal(sm)
				if err != nil {
					fmt.Printf("Error to marshal %s\n", sm)

				} else {
					fmt.Printf("%s\n", jsonStr)
					s := string(jsonStr[:])
					s = "{ \"" + path + "\": " + s + " }"
					status, err := DoPost(a, *a.Config.Giles.GilesServer+"/add", s)
					if err != nil {
						//fmt.Printf("Error Post:%s\n", status)
						MQTT.ERROR.Printf("%sPost Status:%s\n", LORA, status)
						//TODO add to local cache here
					} else if status == 200 {
						a.UUIDS[u1] = true
					}
				}

			} else {
				fmt.Printf("Unknown Sensor\n")
			}
		}
	}
}
