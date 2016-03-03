package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
)

type Category struct {
	Name   string `json:"name"`
	Folder string `json:"folder"`
}

type Categories struct {
	data []*Category
}

func NewCategories() *Categories {
	return &Categories{}
}

func (c *Categories) Load() error {
	return c.LoadFrom("config/categories.json")
}
func (c *Categories) Save() error {
	return c.SaveAs("config/categories.json")
}

func (c *Categories) LoadFrom(filename string) error {
	log.Debugf("Loading categories from %s", filename)

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	jsonParser := json.NewDecoder(file)
	if err = jsonParser.Decode(&c.data); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"addr": fmt.Sprintf("%p", c),
		"len":  len(c.data),
	}).Debug("Categories.LoadFrom")

	return nil
}

func (c *Categories) SaveAs(filename string) error {
	configFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	b, err := json.Marshal(c.data)
	if err != nil {
		return err
	}
	json.Indent(&out, b, "", "    ")
	_, err = out.WriteTo(configFile)
	return err
}

func (c *Categories) GetAll() []*Category {
	log.WithFields(log.Fields{
		"addr": fmt.Sprintf("%p", c),
		"len":  len(c.data),
	}).Debug("Categories.GetAll")
	return c.data
}
