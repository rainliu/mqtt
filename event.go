package mqtt

import "net"

////////////////////Interface//////////////////////////////

type EventType int

const (
	EVENT_SESSION_CREATED EventType = iota
	EVENT_SESSION_TERMINATED

	EVENT_CONNECT
	EVENT_PUBLISH
	EVENT_SUBSCRIBE
	EVENT_UNSUBSCRIBE

	EVENT_TIMEOUT
	EVENT_IOEXCEPTION
)

type TimeoutType int

const (
	TIMEOUT_RETRANSMIT TimeoutType = iota
	TIMEOUT_SESSION
)

type Event interface {
	GetEventType() EventType
	GetSession() Session
}

type EventConnect interface {
	Event

	GetConnectFlags() byte
	GetKeepAlive() uint16
	GetClientId() string
	GetWillTopic() string
	GetWillMessage() string
	GetUserName() string
	GetPassword() []byte
}

type EventPublish interface {
	Event

	GetMessage() Message
}

type EventSubscribe interface {
	Event

	GetSubscribeTopics() []string
	GetQoSs() []QOS
}

type EventUnsubscribe interface {
	Event

	GetUnsubscribeTopics() []string
}

type EventSessionCreated interface {
	Event

	GetReason() string
}

type EventSessionTerminated interface {
	Event

	GetReason() string
}

type EventTimeout interface {
	Event

	GetTimeoutType() TimeoutType
}

type EventIOException interface {
	Event

	GetRemoteAddr() net.Addr
}

////////////////////Implementation////////////////////////
type event struct {
	eventType EventType
	session   Session
}

func newEvent(e EventType, s Session) *event {
	return &event{eventType: e, session: s}
}

func (this *event) GetEventType() EventType {
	return this.eventType
}

func (this *event) GetSession() Session {
	return this.session
}

type event_session struct {
	event

	reason string
}

func newEventSession(e EventType, s Session, r string) *event_session {
	this := &event_session{}

	this.eventType = e
	this.session = s
	this.reason = r

	return this
}

func (this *event_session) GetReason() string {
	return this.reason
}

type event_timeout struct {
	event

	timeoutType TimeoutType
}

func newEventTimeout(e EventType, s Session, t TimeoutType) *event_timeout {
	this := &event_timeout{}

	this.eventType = e
	this.session = s
	this.timeoutType = t

	return this
}

func (this *event_timeout) GetTimeoutType() TimeoutType {
	return this.timeoutType
}

type event_ioexception struct {
	event

	addr net.Addr
}

func newEventIOException(e EventType, s Session, addr net.Addr) *event_ioexception {
	this := &event_ioexception{}

	this.eventType = e
	this.session = s
	this.addr = addr

	return this
}

func (this *event_ioexception) GetRemoteAddr() net.Addr {
	return this.addr
}
