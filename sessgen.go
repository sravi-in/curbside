package curbside

import (
	"io/ioutil"
	"net/http"
	"sync"
)

const getSessionURL = "http://localhost:8000/get-session"

type SessGen struct {
	repeat int
	next   chan string
	stop   chan struct{}
	Sess   chan string
	wg     sync.WaitGroup
}

func NewSessGen(repeat int) *SessGen {
	var sg SessGen
	sg.repeat = repeat
	sg.next = make(chan string)
	sg.stop = make(chan struct{})
	sg.Sess = make(chan string)

	sg.wg.Add(2)
	// Generate new session ID one after other in channel 'next'
	go sg.genNext()
	go sg.genSess()

	return &sg
}

func (sg *SessGen) genNext() {
	defer sg.wg.Done()
	defer close(sg.next)
	for {
		select {
		case <-sg.stop:
			return
		case sg.next <- getSession():
		}
	}
}

func (sg *SessGen) genSess() {
	var sess string
	var ok bool
	i := sg.repeat
	defer sg.wg.Done()
	defer close(sg.Sess)
	for {
		if i < sg.repeat {
			select {
			case sg.Sess <- sess:
				i++
			case <-sg.stop:
				return
			}
		} else {
			select {
			case sess, ok = <-sg.next:
				if !ok {
					return
				}
				i = 0
			case <-sg.stop:
				return
			}
		}
	}
}

func (sg *SessGen) Stop() {
	close(sg.stop)
	sg.wg.Wait()
	// consume channel to prevent leaks
	for range sg.next {
	}
	for range sg.Sess {
	}
}

func getSession() string {
	resp, err := http.Get(getSessionURL)
	if err != nil {
		//return "" //, fmt.Errorf("get session: %v", err)
		panic("get session failed")
	}

	defer resp.Body.Close()
	sess, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "" //, fmt.Errorf("reading rsp from %s: %v", getSessionURL, err)
	}

	return string(sess) //, nil
}
