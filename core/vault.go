package core

import "time"

type command int

const (
	keyInit command = iota
	keyView command = iota
)

type vault struct {
	key        Coffer
	cmdRecv    chan command
	dataReturn chan Buffer
	errReturn  chan error
}

var v = func() (v vault) {
	var err error

	v.key, err = NewCoffer()
	if err != nil {
		errorThenExit(err)
	}

	v.cmdRecv = make(chan command)
	v.dataReturn = make(chan Buffer)
	v.errReturn = make(chan error)

	go func() {
		for cmd := range v.cmdRecv {
			switch cmd {
			case keyInit:
				v.errReturn <- v.key.init()
			case keyView:
				k, err := v.key.view()
				if err != nil {
					v.errReturn <- err
				} else {
					v.dataReturn <- k
				}
			}
		}
	}()

	go func() {
		var err error
		for {
			time.Sleep(interval)

			v.key.mu.Lock()
			err = v.key.rekey()
			v.key.mu.Unlock()

			if err != nil {
				return
			}
		}
	}()

	v.init()

	return
}()

func (v vault) init() (err error) {
	v.key.mu.Lock()
	v.cmdRecv <- keyInit
	err = <-v.errReturn
	v.key.mu.Unlock()
	return
}

func (v vault) view() (d Buffer) {
	v.key.mu.RLock()
	v.cmdRecv <- keyView
	d = <-v.dataReturn
	v.key.mu.RUnlock()
	return
}

func (v vault) destroy() (err error) {
	v.key.mu.Lock()
	err = v.key.destroy()
	v.key.mu.Unlock()
	close(v.cmdRecv)
	close(v.dataReturn)
	close(v.errReturn)
	return
}
