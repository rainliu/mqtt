package mqtt

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

////////////////////Interface//////////////////////////////

type Provider interface {
	AddTransport(t Transport)
	GetTransports() []Transport
	RemoveTransport(t Transport)

	AddListener(l Listener)
	RemoveListener(l Listener)

	Forward(m Message)
}

////////////////////Implementation////////////////////////

type provider struct {
	listeners       map[Listener]Listener
	transports 		map[Transport]Transport
	sessions   		map[Session]*session

	forward   chan Message
	join      chan *session
	leave	  chan *session
	
	quit      chan bool
	waitGroup *sync.WaitGroup
}

func newProvider() *provider {
	this := &provider{}

	this.listeners = make(map[Listener]Listener)
	this.transports = make(map[Transport]Transport)
	this.sessions = make(map[Session]*session)

	this.forward = make(chan Message)
	this.join = make(chan *session)
	this.leave = make(chan *session)
	
	this.quit = make(chan bool)
	this.waitGroup = &sync.WaitGroup{}

	return this
}

func (this *provider) AddTransport(t Transport) {
	this.transports[t] = t
}

func (this *provider) GetTransports() []Transport {
	ts := make([]Transport, len(this.transports))

	l := 0
	for _, t := range this.transports {
		ts[l] = t
		l++
	}

	return ts
}

func (this *provider) RemoveTransport(t Transport) {
	delete(this.transports, t)
}

func (this *provider) AddListener(l Listener) {
	this.listeners[l] = l
}

func (this *provider) RemoveListener(l Listener) {
	delete(this.listeners, l)
}

func (this *provider) Run() {
	for _, t := range this.transports {
		if err := t.Listen(); err != nil {
			log.Printf("Listening %s://%s:%d Failed!!!\n", t.GetNetwork(), t.GetAddress(), t.GetPort())
		} else {
			log.Printf("Listening %s://%s:%d Runing...\n", t.GetNetwork(), t.GetAddress(), t.GetPort())
			this.waitGroup.Add(1)
			go this.ServeAccept(t.(*transport))
		}
	}
	
	//infinite loop run until ctrl+c
	for {
		select {
		case s := <-this.join:
			this.sessions[s] = s
		case s := <-this.leave:
			delete(this.sessions, s)
		case msg := <-this.forward:
			for _, s := range this.sessions {
				if err := s.Forward(msg); err != nil {
					log.Println(err)
					for _, l := range this.listeners {
						l.ProcessIOException(newEventIOException(s, s.conn.RemoteAddr()))
					}
				}
			}
		case <-this.quit:
			log.Println("ServeForward Quit")
			return
		}
	}
}

func (this *provider) Stop() {
	this.quit <- true
	for _, s := range this.sessions {
		s.Terminate(errors.New("Provider Stopped\n"))
	}
	for _, t := range this.transports {
		t.Close()
	}
	this.waitGroup.Wait()
}

func (this *provider) ServeAccept(t *transport) {
	defer this.waitGroup.Done()
	defer t.lner.Close()

	for {
		select {
		case <-t.quit:
			log.Printf("Listening %s://%s:%d Stoped!!!\n", t.GetNetwork(), t.GetAddress(), t.GetPort())
			return
		default:
			//can't delete default, otherwise blocking call
		}
		t.SetDeadline(time.Now().Add(1e9))
		conn, err := t.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); !(ok && opErr.Timeout()) {
				log.Println(err)
			}
			continue
		}
		this.waitGroup.Add(1)
		go this.ServeConn(conn)
	}
}

func (this *provider) ServeConn(conn net.Conn) {
	defer this.waitGroup.Done()
	defer conn.Close()

	s := newSession(conn)
	this.join <- s

	var buf []byte
	var err error
	for {
		select {
		case <-s.quit:
			log.Println("Disconnecting", conn.RemoteAddr())
			for _, l := range this.listeners {
				l.ProcessSessionTerminated(newEventSessionTerminated(s, s.Error(), s.Will()))
			}
			this.leave <- s
			return
		default:
			//can't delete default, otherwise blocking call
		}

		if buf, err = this.ReadPacket(conn); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				s.keepAliveAccumulated += 1 //add 1 second
				if s.keepAlive != 0 && s.keepAliveAccumulated >= (s.keepAlive*3)/2 {
					log.Println("Timeout", conn.RemoteAddr())
					for _, l := range this.listeners {
						l.ProcessTimeout(newEventTimeout(s, TIMEOUT_SESSION))
					}
					s.Terminate(errors.New("Timeout"))
				} else {
					continue
				}
			} else {
				log.Println(err)
				for _, l := range this.listeners {
					l.ProcessIOException(newEventIOException(s, conn.RemoteAddr()))
				}
				s.Terminate(err)
			}
		} else {
			s.keepAliveAccumulated = 0
			if evt := s.Process(buf); evt != nil {
				switch evt.GetEventType() {
				case EVENT_CONNECT:
					for _, l := range this.listeners {
						l.ProcessConnect(evt.(EventConnect))
					}
				case EVENT_PUBLISH:
					for _, l := range this.listeners {
						l.ProcessPublish(evt.(EventPublish))
					}
				case EVENT_SUBSCRIBE:
					for _, l := range this.listeners {
						l.ProcessSubscribe(evt.(EventSubscribe))
					}
				case EVENT_UNSUBSCRIBE:
					for _, l := range this.listeners {
						l.ProcessUnsubscribe(evt.(EventUnsubscribe))
					}
				case EVENT_IOEXCEPTION:
					for _, l := range this.listeners {
						l.ProcessIOException(evt.(EventIOException))
					}
					s.Terminate(errors.New(s.Error()))
				default:
					s.Terminate(errors.New(s.Error()))
				}
			}
		}
	}
}

func (this *provider) Forward(msg Message) {
	this.forward <- msg
}

func (this *provider) ReadPacket(conn net.Conn) ([]byte, error) {
	var pkt [1]byte
	var buf []byte
	var err error

	conn.SetDeadline(time.Now().Add(1e9)) //wait for 1 second
	if _, err = conn.Read(pkt[:]); err != nil {
		return nil, err
	}
	buf = append(buf, pkt[0])

	var remainingLength uint32 = 0
	var multiplier uint32 = 1
	for {
		if _, err = conn.Read(pkt[:]); err != nil {
			return nil, err
		}
		buf = append(buf, pkt[0])
		remainingLength += uint32(pkt[0]&127) * multiplier
		if (pkt[0] & 128) == 0 {
			break
		}
		multiplier *= 128
	}

	if remainingLength > 0 {
		data := make([]byte, remainingLength)
		if _, err = io.ReadFull(conn, data); err != nil {
			return nil, err
		}
		buf = append(buf, data...)
	}

	return buf, nil
}
