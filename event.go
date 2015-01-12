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

type PublishEvent interface {
	Event

	GetMessage() Message
}

type SubscribeEvent interface {
	Event

	GetSubscribeTopics() []string
	GetQos() []QOS
}

type UnsubscribeEvent interface {
	Event

	GetUnsubscribeTopics() []string
}

type SessionCreatedEvent interface {
	Event

	GetReason() string
}

type SessionTerminatedEvent interface {
	Event

	GetReason() string
}

type TimeoutEvent interface {
	Event

	GetTimeoutType() TimeoutType
}

type IOExceptionEvent interface {
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

type sessionEvent struct {
	event

	reason string
}

func newSessionEvent(e EventType, s Session, r string) *sessionEvent {
	this := &sessionEvent{}

	this.eventType = e
	this.session = s
	this.reason = r

	return this
}

func (this *sessionEvent) GetReason() string {
	return this.reason
}

type timeoutEvent struct {
	event

	timeoutType TimeoutType
}

func newTimeoutEvent(e EventType, s Session, t TimeoutType) *timeoutEvent {
	this := &timeoutEvent{}

	this.eventType = e
	this.session = s
	this.timeoutType = t

	return this
}

func (this *timeoutEvent) GetTimeoutType() TimeoutType {
	return this.timeoutType
}

type ioExceptionEvent struct {
	event

	addr net.Addr
}

func newIOExceptionEvent(e EventType, s Session, addr net.Addr) *ioExceptionEvent {
	this := &ioExceptionEvent{}

	this.eventType = e
	this.session = s
	this.addr = addr

	return this
}

func (this *ioExceptionEvent) GetRemoteAddr() net.Addr {
	return this.addr
}
