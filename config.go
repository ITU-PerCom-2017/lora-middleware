package main

import (
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/gcfg.v1"
)

type Config struct {
	TTN struct {
		MQTTServer   *string
		MQTTTopic    *string
		MQTTUsername *string
		MQTTPassword *string
	}

	Giles struct {
		GilesServer   *string
		GilesUsername *string
		GilesPassword *string
	}
}

func LoadConfig(filename string) *Config {
	var configuration Config
	err := gcfg.ReadFileInto(&configuration, filename)
	if err != nil {
		MQTT.DEBUG.Printf("No configuration file found at %v, so checking current directory for lora.cfg (%v)", filename, err)
	} else {
		return &configuration
	}
	err = gcfg.ReadFileInto(&configuration, "./lora.cfg")
	if err != nil {
		MQTT.ERROR.Printf("Could not find configuration files ./lora.cfg. Try retreiving a sample from github.com/ITU-PerCom-2017/lora-middleware")
	} else {
		return &configuration
	}
	return &configuration
}
