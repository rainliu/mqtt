package mqtt

type EventType int

const (
	EVENT_PUBLISH EventType = iota
	EVENT_SUBSCRIBE
	EVENT_UNSUBSCRIBE
	EVENT_TIMEOUT
	EVENT_SESSION_TERMINATED
	EVENT_IOEXCEPTION
)

type Event interface {
	GetEventType() EventType
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

type TimeoutEvent interface {
	Event

	GetTimeout() int
}

type SessionTerminatedEvent interface {
	Event

	GetSession() Session
}

type IOExceptionEvent interface {
	Event

	GetTransport() Transport
}
