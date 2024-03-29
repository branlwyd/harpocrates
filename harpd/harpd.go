package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/BranLwyd/harpocrates/harpd/handler"
	"github.com/BranLwyd/harpocrates/harpd/server"
	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"

	cpb "github.com/BranLwyd/harpocrates/harpd/proto/config_go_proto"
	kpb "github.com/BranLwyd/harpocrates/secret/proto/key_go_proto"
)

var (
	configFile = flag.String("config", "", "The harpd configuration file to use.")
)

// serv implements server.Server.
type serv struct{}

func (serv) ParseConfig() (_ *cpb.Config, _ *kpb.Key, _ error) {
	// Read & parse the config.
	cfgBytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't read config file: %w", err)
	}
	cfg := &cpb.Config{}
	if err := proto.UnmarshalText(string(cfgBytes), cfg); err != nil {
		return nil, nil, fmt.Errorf("couldn't parse config file: %w", err)
	}

	// Fill in sesnsible defaults for some fields if needed.
	if cfg.SessionDurationS == 0 {
		cfg.SessionDurationS = 300
	}
	if cfg.NewSessionRate == 0 {
		cfg.NewSessionRate = 1
	}

	// Sanity check config values.
	if cfg.HostName == "" {
		return nil, nil, errors.New("host_name is required in config")
	}
	if cfg.Email == "" {
		return nil, nil, errors.New("email is required in config")
	}
	if cfg.CertDir == "" {
		return nil, nil, errors.New("cert_dir is required in config")
	}
	if cfg.PassLoc == "" {
		return nil, nil, errors.New("pass_loc is required in config")
	}
	if cfg.KeyFile == "" {
		return nil, nil, errors.New("key_file is required in config")
	}
	if cfg.SessionDurationS <= 0 {
		return nil, nil, errors.New("session_duration_s must be positive")
	}
	if cfg.NewSessionRate <= 0 {
		return nil, nil, errors.New("new_session_rate must be positive")
	}

	if cfg.AlertCmd == "" {
		log.Printf("No alert_cmd specified, logging alerts")
	}

	// Create key, counter store based on config.
	keyBytes, err := ioutil.ReadFile(cfg.KeyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't read key file: %w", err)
	}
	k := &kpb.Key{}
	if err := proto.Unmarshal(keyBytes, k); err != nil {
		return nil, nil, fmt.Errorf("couldn't parse key: %w", err)
	}

	return cfg, k, nil
}

func (serv) Serve(cfg *cpb.Config, h http.Handler) error {
	certMgr := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.HostName),
		Cache:      autocert.DirCache(cfg.CertDir),
		Email:      cfg.Email,
	}
	server := &http.Server{
		TLSConfig: &tls.Config{
			MinVersion:             tls.VersionTLS13,
			SessionTicketsDisabled: true,
			GetCertificate:         certMgr.GetCertificate,
			NextProtos:             []string{"h2", acme.ALPNProto},
		},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handler.NewLogging("https", handler.NewSecureHeader(h)),
	}

	log.Printf("Serving")
	return server.ListenAndServeTLS("", "")
}

func main() {
	flag.Parse()
	if *configFile == "" {
		log.Fatalf("--config is required")
	}
	server.Run(serv{})
}
