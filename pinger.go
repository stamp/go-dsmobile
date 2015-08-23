package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/tatsushid/go-fastping"
)

func pinger(ds *DsMobile) {
	ra, err := net.ResolveIPAddr("ip4:icmp", ds.Ip)
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
				log.Printf("Online, RTT: %v\n", rtt)
				online = true
				ds.Start()
			}
		}
	}
	p.OnIdle = func() {
		if waitingForAnswer {
			waitingForAnswer = false
			if online {
				log.Println("Offline")
				online = false
				ds.Stop()
			}
		}
	}

	for {
		waitingForAnswer = true
		err = p.Run()
		if err != nil {
			fmt.Println(err)
		}
		<-time.After(10 * time.Second)
	}
}
