package main

import "time"

// Copyright © 2017 The Things Network
// UplinkMessage represents an application-layer uplink message
type UplinkMessage struct {
	AppID          string                 `json:"app_id,omitempty"`
	DevID          string                 `json:"dev_id,omitempty"`
	HardwareSerial string                 `json:"hardware_serial,omitempty"`
	FPort          uint8                  `json:"port"`
	FCnt           uint32                 `json:"counter"`
	Confirmed      bool                   `json:"confirmed,omitempty"`
	IsRetry        bool                   `json:"is_retry,omitempty"`
	PayloadRaw     []byte                 `json:"payload_raw"`
	PayloadFields  map[string]interface{} `json:"payload_fields,omitempty"`
	Metadata       Metadata               `json:"metadata,omitempty"`
}

// Metadata contains metadata of a message
type Metadata struct {
	Time       time.Time         `json:"time,omitempty,omitempty"`
	Frequency  float32           `json:"frequency,omitempty"`
	Modulation string            `json:"modulation,omitempty"`
	DataRate   string            `json:"data_rate,omitempty"`
	Bitrate    uint32            `json:"bit_rate,omitempty"`
	CodingRate string            `json:"coding_rate,omitempty"`
	Gateways   []GatewayMetadata `json:"gateways,omitempty"`
	LocationMetadata
}

// LocationMetadata contains GPS coordinates
type LocationMetadata struct {
	Latitude  float32 `json:"latitude,omitempty"`
	Longitude float32 `json:"longitude,omitempty"`
	Altitude  int32   `json:"altitude,omitempty"`
}

// GatewayMetadata contains metadata for each gateway that received a message
type GatewayMetadata struct {
	GtwID      string   `json:"gtw_id,omitempty"`
	GtwTrusted bool     `json:"gtw_trusted,omitempty"`
	Timestamp  uint32   `json:"timestamp,omitempty"`
	Time       JSONTime `json:"time,omitempty"`
	Channel    uint32   `json:"channel"`
	RSSI       float32  `json:"rssi,omitempty"`
	SNR        float32  `json:"snr,omitempty"`
	RFChain    uint32   `json:"rf_chain,omitempty"`
	LocationMetadata
}

// JSONTime is a time.Time that marshals to/from RFC3339Nano format
type JSONTime time.Time

// MarshalText implements the encoding.TextMarshaler interface
func (t JSONTime) MarshalText() ([]byte, error) {
	if time.Time(t).IsZero() || time.Time(t).Unix() == 0 {
		return []byte{}, nil
	}
	stamp := time.Time(t).UTC().Format(time.RFC3339Nano)
	return []byte(stamp), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (t *JSONTime) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*t = JSONTime{}
		return nil
	}
	time, err := time.Parse(time.RFC3339Nano, string(text))
	if err != nil {
		return err
	}
	*t = JSONTime(time)
	return nil
}

// BuildTime builds a new JSONTime
func BuildTime(unixNano int64) JSONTime {
	if unixNano == 0 {
		return JSONTime{}
	}
	return JSONTime(time.Unix(0, 0).Add(time.Duration(unixNano)).UTC())
}

type SmapMeta struct {
	UUID       string     `json:"uuid"`
	Properties *Property  `json:",omitempty"`
	Path       string     `json:",omitempty"` //use path as resource path
	Metadata   *Metadatum `json:",omitempty"`
	//Readings   []string  `json:"Readings"`
	Readings [][]float64 `json:readings`
	//Actuator   Actuator  `json:",omitempty"`
}

type Property struct {
	Timezone      string `json:",omitempty"`
	UnitofMeasure string `json:",omitempty"`
	ReadingType   string `json:",omitempty"`
}

type Metadatum struct {
	SourceName string      `json:",omitempty"`
	Instrument Instrument  `json:",omitempty"`
	Location   Location    `json:",omitempty"`
	System     string      `json:",omitempty"`
	Extra      interface{} `json:",omitempty"` //NOTE using interface because extra has no pre-known structure
}

type Instrument struct {
	Manufacturer   string `json:",omitempty"`
	Model          string `json:",omitempty"`
	SamplingPeriod string `json:",omitempty"`
}

type Location struct {
	City        string `json:",omitempty"`
	Building    string `json:",omitempty"`
	Campus      string `json:",omitempty"`
	Floor       string `json:",omitempty"`
	Section     string `json:",omitempty"`
	Room        string `json:",omitempty"`
	Coordinates string `json:", omitempty"`
}

type Actuator struct {
	States   string `json:",omitempty"`
	Values   string `json:",omitempty"`
	Model    string `json:",omitempty"`
	MinValue string `json:",omitempty"`
	MaxValue string `json:",omitempty"`
}

// FIXME how to represent readings?
type Reading struct {
	UUID    string    `json:uuid`
	Time    time.Time `json:time`
	Reading float64   `json:reading`
}

// we just care about reading and UUID
type SmapReading struct {
	//Resource string      `json:resource`
	Readings [][]float64 `json:readings`
	UUID     string      `json:"uuid"`
}
