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
	//"log"

	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/satori/go.uuid"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	go giles_pup(message)
}

var i int64

var NS uuid.UUID
var UUIDS map[string]bool

func init() {
	MQTT.DEBUG = log.New(os.Stdout, "DEBUG ", 0)
	MQTT.WARN = log.New(os.Stdout, "WARN ", 0)
	MQTT.CRITICAL = log.New(os.Stdout, "CRITICAL ", 0)
	MQTT.ERROR = log.New(os.Stdout, "ERROR ", 0)
}

func main() {
	//msg := "ABC"
	//encoded := base64.StdEncoding.EncodeToString([]byte(msg))
	//fmt.Println(encoded)
	//encoded = "VDIwSDYw"
	//decoded, err := base64.StdEncoding.DecodeString(encoded)
	//if err != nil {
	//fmt.Println("decode error:", err)
	//return
	//}
	//fmt.Println(string(decoded))
	UUIDS = make(map[string]bool)
	var err error
	NS, err = uuid.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		fmt.Printf("Something gone wrong: %s", err)
		os.Exit(0)
	}

	//MQTT.DEBUG = log.New(os.Stdout, "", 0)
	//MQTT.ERROR = log.New(os.Stdout, "", 0)
	c := make(chan os.Signal, 1)
	i = 0
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("signal received, exiting")
		os.Exit(0)
	}()

	hostname, _ := os.Hostname()

	server := flag.String("server", "tcp://127.0.0.1:1883", "The full url of the MQTT server to connect to ex: tcp://127.0.0.1:1883")
	topic := flag.String("topic", "#", "Topic to subscribe to")
	qos := flag.Int("qos", 0, "The QoS to subscribe to messages at")
	clientid := flag.String("clientid", hostname+strconv.Itoa(time.Now().Second()), "A clientid for the connection")
	username := flag.String("username", "", "A username to authenticate to the MQTT server")
	password := flag.String("password", "", "Password to match username")
	flag.Parse()

	connOpts := &MQTT.ClientOptions{
		ClientID:             *clientid,
		CleanSession:         true,
		Username:             *username,
		Password:             *password,
		MaxReconnectInterval: 1 * time.Second,
		//KeepAlive:            60 * time.Second,
		TLSConfig: tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert},
	}
	connOpts.AddBroker(*server)
	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(*topic, byte(*qos), onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to %s\n", *server)
	}

	for {
		time.Sleep(100 * time.Millisecond)
	}
}

func giles_pup(message MQTT.Message) {
	var m UplinkMessage
	s := message.Payload()

	//fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
	if err := json.Unmarshal(s, &m); err != nil {
		fmt.Printf("Error with %s\n", s)
	} else {
		j, err := json.MarshalIndent(m, "", " ")
		if err != nil {
			fmt.Printf("Error to marshal %s\n", m)

		} else if len(m.PayloadRaw) == 3 {

			fmt.Printf("%v\n", m.PayloadRaw)
			fmt.Printf("%q\n", m.PayloadRaw)
			var sm SmapMeta
			var property Property
			var metadatum Metadatum
			var instrument Instrument
			var location Location

			var sensorType string
			var units string
			var found bool
			var v float64
			// temperature
			model := "Chirp"
			if m.PayloadRaw[0] == 170 {
				sensorType = "Temperature"
				units = "C"
				found = true
				v = float64((int(m.PayloadRaw[1])<<8)+int(m.PayloadRaw[2])) / 10
			} else if m.PayloadRaw[0] == 187 {
				sensorType = "Humidity"
				units = "Moisture"
				found = true
				v = float64((int(m.PayloadRaw[1]) << 8) + int(m.PayloadRaw[2]))
			} else if m.PayloadRaw[0] == 204 { // Light
				sensorType = "Light"
				units = "Light"
				found = true
				v = float64((int(m.PayloadRaw[1]) << 8) + int(m.PayloadRaw[2]))
				v = 100 * ((65535 - v) / 65535)
			}
			if found {
				u1 := uuid.NewV5(NS, m.DevID+model+sensorType).String()
				fmt.Println(u1)
				fmt.Printf("%v\n", v)
				fmt.Printf("%s\n", j)

				tr := make([][]float64, 1)
				tr[0] = make([]float64, 2)
				tr[0][0] = float64(m.Metadata.Time.Unix())
				tr[0][1] = float64(v)

				path := "ITU/5/5A56/" + m.DevID + "/" + model + "/" + sensorType

				_, exists := UUIDS[u1]

				if exists {
					// just post new reading
					sm.UUID = u1
					sm.Readings = tr

				} else {
					metadatum.Instrument = instrument
					metadatum.Location = location
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
					sm.Metadata.Location.Floor = "5"
					sm.Metadata.Location.Room = "5A56"
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
					status, err := DoPost("http://130.226.142.195:8079/add/lopy", s)
					if err != nil {
						fmt.Printf("Error Post:%s\n", status)
						//TODO add to local cache here
					} else {
						UUIDS[u1] = true
					}
				}

			} else {
				fmt.Printf("Unknown Sensor\n")
			}
		}
	}
}
