package gui

import "sync"

type Progress struct {
	Title    string
	Progress float64

	trigger chan struct{}
	done    chan struct{}

	sync.Mutex
}

func NewProgress(title string) *Progress {
	p := &Progress{
		Title:   title,
		trigger: make(chan struct{}),
		done:    make(chan struct{}),
	}

	go p.monitor()

	return p
}

func (self *Progress) Update(progress float64) {
	self.Lock()
	self.Progress = progress
	self.Unlock()

	// Try to send a trigger
	select {
	case self.trigger <- struct{}{}:
	default:

	}
}

func (self *Progress) Done() {
	self.done <- struct{}{}
}

func (self *Progress) monitor() {
	for {
		select {
		case <-self.done:
			// Done :)
			return
		case <-self.trigger:
			// Update

			// Block so updates is triggerd to offen
			select {
			case <-time.After(time.Milisecond * 100):
			case <-self.done:
				// Done :)
				return
			}
		}
	}
}
