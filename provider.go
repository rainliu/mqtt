package mqtt

import (
	"log"
	"net"
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
}

func newProvider() *provider {
	this := &provider{}

	this.listeners = make(map[Listener]Listener)
	this.serverTransports = make(map[ServerTransport]ServerTransport)
	this.clientSessions = make(map[ClientSession]*clientSession)
	this.serverSessions = make(map[ServerSession]*serverSession)

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
			log.Printf("Listening %s//%s:%d Failed!\n", st.GetNetwork(), st.GetAddress(), st.GetPort())
		} else {
			log.Printf("Listening %s//%s:%d ...\n", st.GetNetwork(), st.GetAddress(), st.GetPort())
			go this.ServeAccept(st)
		}
	}
}

func (this *provider) ServeAccept(st ServerTransport) {
	//defer s.waitGroup.Done()
	for {
		// select {
		// case <-s.ch:
		// 	log.Println("stopping listening on", listener.Addr())
		// 	listener.Close()
		// 	return
		// default:
		// }
		// listener.SetDeadline(time.Now().Add(1e9))
		conn, err := st.Accept()
		if err != nil {
			continue
			// if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			// 	continue
			// }
			// log.Println(err)
		}
		//s.waitGroup.Add(1)
		go this.ServeConn(conn)
	}
}

func (this *provider) ServeConn(conn net.Conn) {
	defer conn.Close()

	ss := newServerSession(conn)
	this.serverSessions[ss] = ss
	ss.Run()

	//defer s.waitGroup.Done()
	// for {
	// 	// select {
	// 	// case <-s.ch:
	// 	// 	log.Println("disconnecting", conn.RemoteAddr())
	// 	// 	return
	// 	// default:
	// 	// }
	// 	// conn.SetDeadline(time.Now().Add(1e9))
	// 	buf := make([]byte, 4096)
	// 	if _, err := conn.Read(buf); nil != err {
	// 		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
	// 			continue
	// 		}
	// 		log.Println(err)
	// 		return
	// 	}
	// 	if _, err := conn.Write(buf); nil != err {
	// 		log.Println(err)
	// 		return
	// 	}
	// }
}
