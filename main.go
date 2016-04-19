package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"time"
)

type AppConfig struct {
	ListenAddress  string            `json:"listen"`
	DefaultAddress string            `json:"default"`
	ProbeTimeout   time.Duration     `json:"timeout"`
	ConnectTimeout time.Duration     `json:"connect_timeout"`
	ProtocolList   []json.RawMessage `json:"protocols"`
}

func createProtocol(data json.RawMessage) (Protocol, error) {
	var d interface{}
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}

	var service string
	if m, ok := d.(map[string]interface{}); !ok {
		return nil, errors.New("error config file format")
	} else {
		val, ok := m["service"]
		if !ok {
			return nil, errors.New("service required")
		}

		if service, ok = val.(string); !ok {
			return nil, errors.New("service must be string")
		}
	}

	var cfg ProtocolConfig

	switch service {
	case "mqtt":
		cfg = new(MQTTConfig)
	case "ssh":
		cfg = new(SSHConfig)
	default:
		return nil, errors.New("invalid protocol: " + service)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return cfg.NewProtocol(), nil
}

func parseConfig() AppConfig {
	configFile := flag.String("config", "default.cfg", "Json format config file")
	flag.Parse()

	var config AppConfig
	if f, err := os.Open(*configFile); err != nil {
		log.Fatalf("[FATAL] config file open error:", err)
	} else {
		defer f.Close()
		if err = json.NewDecoder(f).Decode(&config); err != nil {
			log.Fatalln("[FATAL] config parser error")
		}
	}

	return config
}

func main() {
	config := parseConfig()

	mux := NewMux()
	mux.defaultAddress = config.DefaultAddress
	if t := config.ConnectTimeout; t > 0 {
		mux.connectTimeout = t * time.Second
	}
	if t := config.ProbeTimeout; t > 0 {
		mux.probeTimeout = t * time.Second
	}

	for _, rawData := range config.ProtocolList {
		protocol, err := createProtocol(rawData)
		if err != nil {
			log.Panicln("config file error:", err)
		}
		mux.Handle(protocol)
	}

	log.Printf("[INFO] listen: %s\n", config.ListenAddress)

	if err := mux.ListenAndServe(config.ListenAddress); err != nil {
		log.Fatalf("[FATAL] listen: %s\n", err)
	}
}
