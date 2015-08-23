package main

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/satori/go.uuid"
)

type XmlPicture struct {
	NAME     string
	FPATH    string
	EXT      string
	SIZE     string
	TYPE     string
	TIMECODE string
	TIME     string
}

type XmlAllFile struct {
	Picture XmlPicture
}

type XmlList struct {
	ALLFile []XmlAllFile
}

type DsMobile struct {
	Ip      string
	NewFile chan *uuid.UUID

	address  string
	username string
	password string

	stop chan bool
}

func NewDsMobile(ip, username, password string) *DsMobile {
	return &DsMobile{
		Ip:       ip,
		NewFile:  make(chan *uuid.UUID),
		address:  "http://" + ip,
		username: username,
		password: password,

		stop: make(chan bool, 1),
	}
}

func (ds *DsMobile) Start() {
	go ds.Loop()
}
func (ds *DsMobile) Stop() {
	select {
	case ds.stop <- true:
		// Sent stop to run loop, everything worked great
	case <-time.After(time.Second):
		log.Println("Failed to send stop to run loop")
	}
}

func (ds *DsMobile) Loop() {
	t := time.Second * 0

	log.Println("Started loop")
	for {
		select {
		case <-time.After(t):
			t = time.Second * 2

			err := ds.downloadAll()
			if err != nil {
				log.Println(err)
			}
		case <-ds.stop:
			log.Println("Stopped loop")
			return
		}
	}
}

func (ds *DsMobile) downloadAll() error {
	files, err := ds.getFileList(ds.address + "/:sda1/DCIM:.xml:Picture:Sub")
	if err != nil {
		return err
	}

	for _, allFile := range *files {
		picture := allFile.Picture

		uuid, err := ds.downloadFile(ds.address+picture.FPATH, "storage/")
		if err != nil {
			return err
		}

		err = ds.deleteFile(ds.address+"/.xmlalbum.page_index=1.chipsipcmd", picture.FPATH)
		if err != nil {
			return err
		}

		// Try to notify of the new file
		select {
		case ds.NewFile <- uuid:
			// Notification successful
		default:
			// No one is listening, to bad :/
		}
	}

	return nil
}

func (ds *DsMobile) getFileList(url string) (*[]XmlAllFile, error) {
	respBody, err := ds.get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(respBody)
	if err != nil {
		return nil, err
	}

	var l XmlList
	err = xml.Unmarshal(body, &l)
	if err != nil {
		return nil, err
	}

	return &l.ALLFile, nil
}

func (ds *DsMobile) downloadFile(url, destination string) (*uuid.UUID, error) {
	body, err := ds.get(url)
	if err != nil {
		return nil, err
	}

	id := uuid.NewV4()
	filename := destination + "/" + id.String() + ".JPG"

	_, err = os.Stat(filename)
	for err == nil {
		id = uuid.NewV4()
		filename = destination + "/" + id.String() + ".JPG"
		_, err = os.Stat(filename)
	}

	out, err := os.Create(filename)
	defer out.Close()
	n, err := io.Copy(out, body)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, fmt.Errorf("0 bytes have been copied")
	}

	<-time.After(time.Millisecond * 100)
	return &id, nil
}

func (ds *DsMobile) deleteFile(url, file string) error {
	client := &http.Client{}
	form := &bytes.Buffer{}
	writer := multipart.NewWriter(form)
	writer.WriteField("multiDelete", "Delete Select File")
	writer.WriteField("currentPage", "Current_page=[3]")
	writer.WriteField("multiUploadToFB", "")
	writer.WriteField(file, "on")
	writer.Close()

	req, _ := http.NewRequest("POST", url, form)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(ds.username+":"+ds.password)))
	client.Do(req)

	<-time.After(time.Millisecond * 100)
	return nil
}

func (ds *DsMobile) get(url string) (io.Reader, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(ds.username+":"+ds.password)))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
