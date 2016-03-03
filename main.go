package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"./sources"
	"github.com/facebookgo/inject"
	"github.com/koding/multiconfig"
	"github.com/stamp/go-dsmobile/gui"
	"github.com/stamp/go-dsmobile/types"

	log "github.com/Sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//	log.SetFormatter(&log.JSONFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

type Config struct {
	WebPort  string `default:"3001"`
	Ip       string `default:"10.10.100.1"`
	Username string
	Password string
}

type Saveable interface {
	Save() error
}

type Loadable interface {
	Load() error
}

type Startable interface {
	Start() error
}

func main() {
	// Load the main config
	config := &Config{}
	m := multiconfig.NewWithPath("config/config.json")
	m.Load(config)
	saveConfigToFile(config)

	log.Info("Creating services")
	services := make([]interface{}, 0)
	services = append(services, gui.NewWebserver(config.WebPort))
	categories := types.NewCategories()
	services = append(services, categories)

	// DS mobile source
	ds := sources.NewDsMobile(config.Ip, config.Username, config.Password)
	services = append(services, ds)
	// Start the ping monitor
	services = append(services, sources.NewPinger(ds))

	// Start the hub for websocket messages
	hub := gui.NewHub()
	services = append(services, hub)

	err := inject.Populate(services...)
	if err != nil {
		panic("Failed to populate: " + err.Error())
	}

	for _, s := range services {
		log.WithFields(log.Fields{
			"addr": fmt.Sprintf("%p", s),
		}).Debugf("- %T", s)
	}

	// First load everything
	log.Info("Loading data")
	for _, s := range services {
		if s, ok := s.(Loadable); ok {
			if err := s.Load(); err != nil {
				log.Errorf("Failed load %T: %s", s, err)
			}

			log.WithFields(log.Fields{
				"addr": fmt.Sprintf("%p", s),
			}).Debugf("- %T", s)
		}
	}

	// Start all services
	log.Info("Starting services")
	for _, s := range services {
		if s, ok := s.(Startable); ok {
			if err := s.Start(); err != nil {
				log.Errorf("Failed start %T: %s", s, err)
			}
			log.WithFields(log.Fields{
				"addr": fmt.Sprintf("%p", s),
			}).Debugf("- %T", s)
		}
	}

	go func() {
		for {
			select {
			case uuid := <-ds.NewFile:
				type newFileEvent struct {
					Type string
					Uuid string
					Path string
				}

				e := newFileEvent{
					Type: "newFile",
					Uuid: uuid.String(),
					Path: "/storage/" + uuid.String() + ".JPG",
				}

				msg, _ := json.Marshal(e)
				hub.Broadcast <- msg
			}

		}
	}()

	select {}
}

func saveConfigToFile(config *Config) error {
	configFile, err := os.Create("config/config.json")
	if err != nil {
		return err
	}

	var out bytes.Buffer
	b, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}
	json.Indent(&out, b, "", "\t")
	out.WriteTo(configFile)

	return nil
}
