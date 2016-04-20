package main

import (
	"encoding/json"
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

	var pm ProtocolManager
	if err := pm.LoadProtocols(config.ProtocolList); err != nil {
		log.Panicln("protocols load error:", err)
	}

	mux.pm = &pm

	log.Printf("[INFO] listen: %s\n", config.ListenAddress)

	if err := mux.ListenAndServe(config.ListenAddress); err != nil {
		log.Fatalf("[FATAL] listen: %s\n", err)
	}
}
