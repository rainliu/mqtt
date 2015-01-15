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

	CreateClientSession(ct ClientTransport) ClientSession
	GetClientSessions() []ClientSession
	DeleteClientSession(cs ClientSession)

	Forward(msg Message)
}

////////////////////Implementation////////////////////////

type provider struct {
	listeners        map[Listener]Listener
	serverTransports map[ServerTransport]ServerTransport
	clientSessions   map[ClientSession]*clientSession
	serverSessions   map[ServerSession]*serverSession

	ch        chan Message
	waitGroup *sync.WaitGroup
}

func newProvider() *provider {
	this := &provider{}

	this.listeners = make(map[Listener]Listener)
	this.serverTransports = make(map[ServerTransport]ServerTransport)
	this.clientSessions = make(map[ClientSession]*clientSession)
	this.serverSessions = make(map[ServerSession]*serverSession)

	this.waitGroup = &sync.WaitGroup{}

	return this
}

func (this *provider) AddServerTransport(st ServerTransport) {
	this.serverTransports[st] = st
}

func (this *provider) GetServerTransports() []ServerTransport {
	serverTransports := make([]ServerTransport, len(this.serverTransports))

	l := 0
	for _, value := range this.serverTransports {
		serverTransports[l] = value
		l++
	}

	return serverTransports
}

func (this *provider) RemoveServerTransport(st ServerTransport) {
	delete(this.serverTransports, st)
}

func (this *provider) AddListener(l Listener) {
	this.listeners[l] = l
}

func (this *provider) RemoveListener(l Listener) {
	delete(this.listeners, l)
}

func (this *provider) CreateClientSession(ct ClientTransport) ClientSession {
	return nil
}

func (this *provider) GetClientSessions() []ClientSession {
	clientSessions := make([]ClientSession, len(this.clientSessions))

	l := 0
	for _, value := range this.clientSessions {
		clientSessions[l] = value
		l++
	}

	return clientSessions
}

func (this *provider) DeleteClientSession(cs ClientSession) {
	delete(this.clientSessions, cs)
}

func (this *provider) Run() {
	for _, st := range this.serverTransports {
		if err := st.Listen(); err != nil {
			log.Printf("Listening %s://%s:%d Failed!!!\n", st.GetNetwork(), st.GetAddress(), st.GetPort())
		} else {
			log.Printf("Listening %s://%s:%d Runing...\n", st.GetNetwork(), st.GetAddress(), st.GetPort())
			this.waitGroup.Add(1)
			go this.ServeAccept(st.(*transport))
		}
	}
}

func (this *provider) Stop() {
	for _, st := range this.serverTransports {
		st.Close()
	}
	for _, ss := range this.serverSessions {
		ss.Terminate(errors.New("Provider Stopped\n"))
	}
	this.waitGroup.Wait()
}

func (this *provider) ServeAccept(st *transport) {
	defer this.waitGroup.Done()
	for {
		select {
		case <-st.ch:
			log.Printf("Listening %s://%s:%d Stoped!!!\n", st.GetNetwork(), st.GetAddress(), st.GetPort())
			return
		default:
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
	defer conn.Close()
	defer this.waitGroup.Done()

	ss := newServerSession(conn)
	this.serverSessions[ss] = ss

	var buf []byte
	var err error
	for {
		select {
		case <-ss.ch:
			log.Println("Disconnecting", conn.RemoteAddr())
			for _, ln := range this.listeners {
				ln.ProcessSessionTerminated(newEventSessionTerminated(ss, ss.Error()))
			}
			delete(this.serverSessions, ss)
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
				case EVENT_TIMEOUT:
					for _, ln := range this.listeners {
						ln.ProcessTimeout(evt.(EventTimeout))
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

//Change to FanIn channel
func (this *provider) Forward(msg Message) {
	for _, ss := range this.serverSessions {
		if err := ss.Forward(msg); err != nil {
			log.Println(err)
			for _, ln := range this.listeners {
				ln.ProcessIOException(newEventIOException(ss, ss.conn.RemoteAddr()))
			}
		}
	}
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
