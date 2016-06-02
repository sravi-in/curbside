package curbside

import (
	"io/ioutil"
	"net/http"
	"sync"
)

//const getSessionURL = "http://challenge.shopcurbside.com/get-session"

const getSessionURL = "http://localhost:8000/get-session"

type SessGen struct {
	repeat int
	next   chan SessRsp
	stop   chan struct{}
	Sess   chan SessRsp
	sync.WaitGroup
}

type SessRsp struct {
	sess string
	err  error
}

func NewSessGen(repeat int) *SessGen {
	const numNextWorkers = 3
	var sg SessGen
	sg.repeat = repeat
	sg.next = make(chan SessRsp)
	sg.stop = make(chan struct{})
	sg.Sess = make(chan SessRsp)

	sg.Add(numNextWorkers + 1)
	// Generate and keep upto next 3 session IDs
	for i := 0; i < numNextWorkers; i++ {
		go sg.genNext()
	}
	go sg.genSess()

	return &sg
}

func (sg *SessGen) genNext() {
	defer sg.Done()
	for {
		sess, err := getSession()
		select {
		case <-sg.stop:
			return
		case sg.next <- SessRsp{sess, err}:
		}
	}
}

func (sg *SessGen) genSess() {
	var rsp SessRsp
	var ok bool
	i := sg.repeat
	defer sg.Done()
	for {
		select {
		case rsp, ok = <-next:
			if !ok {
				return
			}
			i = 0
			current = sg.Sess
			next = nil
		case current <- rsp:
			i++
			if i == sg.repeat {
				current = nil
				next = sg.next
			}
		case <-sg.stop:
			return
		}
	}
}

func (sg *SessGen) Stop() {
	close(sg.stop)
	sg.Wait()
	close(sg.next)
	close(sg.Sess)
	// consume channel to prevent leaks
	for range sg.next {
	}
	for range sg.Sess {
	}
}

func getSession() (string, error) {
	resp, err := http.Get(getSessionURL)
	if err != nil {
		return "", fmt.Errorf("get session: %v", err)
	}

	defer resp.Body.Close()
	sess, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading rsp from %s: %v", getSessionURL, err)
	}

	return string(sess), nil
}
