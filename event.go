package mqtt

////////////////////Interface//////////////////////////////

type EventType int

const (
	EVENT_PUBLISH EventType = iota
	EVENT_SUBSCRIBE
	EVENT_UNSUBSCRIBE

	EVENT_SESSION_CREATED
	EVENT_SESSION_TERMINATED

	EVENT_TIMEOUT
	EVENT_IOEXCEPTION
)

type Timeout int

const (
	TIMEOUT_RETRANSMIT Timeout = iota
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

	GetTimeout() Timeout
}

type IOExceptionEvent interface {
	Event

	GetTransport() Transport
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
