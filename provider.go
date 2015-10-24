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
	AddServerTransport(st ServerTransport)
	GetServerTransports() []ServerTransport
	RemoveServerTransport(st ServerTransport)

	AddListener(listener Listener)
	RemoveListener(listener Listener)

	Forward(msg Message)
}

////////////////////Implementation////////////////////////

type provider struct {
	listeners       map[Listener]Listener
	transports 		map[ServerTransport]ServerTransport
	clients   		map[ServerSession]*serverSession

	forward   chan Message
	join      chan *serverSession
	leave	  chan *serverSession
	
	quit      chan bool
	waitGroup *sync.WaitGroup
}

func newProvider() *provider {
	this := &provider{}

	this.listeners = make(map[Listener]Listener)
	this.transports = make(map[ServerTransport]ServerTransport)
	this.clients = make(map[ServerSession]*serverSession)

	this.forward = make(chan Message)
	this.join = make(chan *serverSession)
	this.leave = make(chan *serverSession)
	
	this.quit = make(chan bool)
	this.waitGroup = &sync.WaitGroup{}

	return this
}

func (this *provider) AddServerTransport(st ServerTransport) {
	this.transports[st] = st
}

func (this *provider) GetServerTransports() []ServerTransport {
	serverTransports := make([]ServerTransport, len(this.transports))

	l := 0
	for _, value := range this.transports {
		serverTransports[l] = value
		l++
	}

	return serverTransports
}

func (this *provider) RemoveServerTransport(st ServerTransport) {
	delete(this.transports, st)
}

func (this *provider) AddListener(l Listener) {
	this.listeners[l] = l
}

func (this *provider) RemoveListener(l Listener) {
	delete(this.listeners, l)
}

func (this *provider) Run() {
	for _, st := range this.transports {
		if err := st.Listen(); err != nil {
			log.Printf("Listening %s://%s:%d Failed!!!\n", st.GetNetwork(), st.GetAddress(), st.GetPort())
		} else {
			log.Printf("Listening %s://%s:%d Runing...\n", st.GetNetwork(), st.GetAddress(), st.GetPort())
			this.waitGroup.Add(1)
			go this.ServeAccept(st.(*transport))
		}
	}
	
	//infinite loop run until ctrl+c
	for {
		select {
		case client := <-this.join:
			this.clients[client] = client
		case client := <-this.leave:
			delete(this.clients, client)
		case msg := <-this.forward:
			for _, client := range this.clients {
				if err := client.Forward(msg); err != nil {
					log.Println(err)
					for _, ln := range this.listeners {
						ln.ProcessIOException(newEventIOException(client, client.conn.RemoteAddr()))
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
	for _, ss := range this.clients {
		ss.Terminate(errors.New("Provider Stopped\n"))
	}
	for _, st := range this.transports {
		st.Close()
	}
	this.waitGroup.Wait()
}

func (this *provider) ServeAccept(st *transport) {
	defer this.waitGroup.Done()
	defer st.lner.Close()

	for {
		select {
		case <-st.quit:
			log.Printf("Listening %s://%s:%d Stoped!!!\n", st.GetNetwork(), st.GetAddress(), st.GetPort())
			return
		default:
			//can't delete default, otherwise blocking call
		}
		st.SetDeadline(time.Now().Add(1e9))
		conn, err := st.Accept()
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

	ss := newServerSession(conn)
	this.join <- ss

	var buf []byte
	var err error
	for {
		select {
		case <-ss.quit:
			log.Println("Disconnecting", conn.RemoteAddr())
			for _, ln := range this.listeners {
				ln.ProcessSessionTerminated(newEventSessionTerminated(ss, ss.Error(), ss.Will()))
			}
			this.leave <- ss
			return
		default:
			//can't delete default, otherwise blocking call
		}

		if buf, err = this.ReadPacket(conn); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				ss.keepAliveAccumulated += 1 //add 1 second
				if ss.keepAlive != 0 && ss.keepAliveAccumulated >= (ss.keepAlive*3)/2 {
					log.Println("Timeout", conn.RemoteAddr())
					for _, ln := range this.listeners {
						ln.ProcessTimeout(newEventTimeout(ss, TIMEOUT_SESSION))
					}
					ss.Terminate(errors.New("Timeout"))
				} else {
					continue
				}
			} else {
				log.Println(err)
				for _, ln := range this.listeners {
					ln.ProcessIOException(newEventIOException(ss, conn.RemoteAddr()))
				}
				ss.Terminate(err)
			}
		} else {
			ss.keepAliveAccumulated = 0
			if evt := ss.Process(buf); evt != nil {
				switch evt.GetEventType() {
				case EVENT_CONNECT:
					for _, ln := range this.listeners {
						ln.ProcessConnect(evt.(EventConnect))
					}
				case EVENT_PUBLISH:
					for _, ln := range this.listeners {
						ln.ProcessPublish(evt.(EventPublish))
					}
				case EVENT_SUBSCRIBE:
					for _, ln := range this.listeners {
						ln.ProcessSubscribe(evt.(EventSubscribe))
					}
				case EVENT_UNSUBSCRIBE:
					for _, ln := range this.listeners {
						ln.ProcessUnsubscribe(evt.(EventUnsubscribe))
					}
				case EVENT_IOEXCEPTION:
					for _, ln := range this.listeners {
						ln.ProcessIOException(evt.(EventIOException))
					}
					ss.Terminate(errors.New(ss.Error()))
				default:
					ss.Terminate(errors.New(ss.Error()))
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
