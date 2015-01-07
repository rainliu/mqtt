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

	GetTopics() []string
}

type UnsubscribeEvent interface {
	Event

	GetTopics() []string
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
