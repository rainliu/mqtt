package mqtt

import (
	"errors"
	"io/ioutil"
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
	DeleteClientSession(ct ClientSession)

	Forward(m Message)
}

////////////////////Implementation////////////////////////

type provider struct {
	listeners        map[Listener]Listener
	serverTransports map[ServerTransport]ServerTransport
	clientSessions   map[ClientSession]*clientSession
	serverSessions   map[ServerSession]*serverSession

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

func (this *provider) DeleteClientSession(c ClientSession) {
	delete(this.clientSessions, c)
}

func (this *provider) Forward(m Message) {
	for _, st := range this.serverSessions {
		st.Forward(m)
	}
}

func (this *provider) Run() {
	for _, st := range this.serverTransports {
		if err := st.Listen(); err != nil {
			log.Printf("Listening %s//%s:%d Failed!!!\n", st.GetNetwork(), st.GetAddress(), st.GetPort())
		} else {
			log.Printf("Listening %s//%s:%d Runing...\n", st.GetNetwork(), st.GetAddress(), st.GetPort())
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
			log.Println("Listening %s//%s:%d Stoped!!!\n", st.GetNetwork(), st.GetAddress(), st.GetPort())
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
	for _, ln := range this.listeners {
		ln.ProcessSessionCreated(newSessionEvent(EVENT_SESSION_CREATED, ss, "New Connection Coming"))
	}

	var packet []byte
	var err error
	for {
		select {
		case <-ss.ch:
			log.Println("disconnecting", conn.RemoteAddr())
			break
		default:
		}
		conn.SetDeadline(time.Now().Add(1e9))
		if packet, err = ioutil.ReadAll(conn); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Println(err)
			ss.Terminate(err)
		} else {
			ss.Process(packet)
		}
	}

	for _, ln := range this.listeners {
		ln.ProcessSessionTerminated(newSessionEvent(EVENT_SESSION_TERMINATED, ss, ss.Error()))
	}
	delete(this.serverSessions, ss)
}
