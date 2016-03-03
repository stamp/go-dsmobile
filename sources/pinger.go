package sources

import (
	"fmt"
	"net"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tatsushid/go-fastping"
)

type Pinger struct {
	ds *DsMobile
}

func NewPinger(ds *DsMobile) *Pinger {
	return &Pinger{ds: ds}
}

func (self *Pinger) Start() error {
	go func() {
		ra, err := net.ResolveIPAddr("ip4:icmp", self.ds.Ip)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		waitingForAnswer := false
		online := false
		p := fastping.NewPinger()
		p.AddIPAddr(ra)

		p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
			if waitingForAnswer {
				waitingForAnswer = false
				if !online {
					log.WithFields(log.Fields{
						"rtt": rtt,
					}).Info("Pinger: Online")
					online = true
					self.ds.Start()
				}
			}
		}
		p.OnIdle = func() {
			if waitingForAnswer {
				waitingForAnswer = false
				if online {
					log.Info("Pinger: Offline")
					online = false
					self.ds.Stop()
				}
			}
		}

		for {
			waitingForAnswer = true

			err = p.Run()
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Errorf("Ping %s failed", ra.String())
			}
			<-time.After(10 * time.Second)
		}
	}()
	return nil
}
