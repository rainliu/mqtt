package mqtt

////////////////////Interface//////////////////////////////

type Listener interface {
	ProcessSessionCreated(eventSessionCreated EventSessionCreated) error
	ProcessSessionTerminated(eventSessionTerminated EventSessionTerminated) error

	ProcessConnect(eventConnect EventConnect) error
	ProcessPublish(eventPublish EventPublish) error
	ProcessSubscribe(eventSubscribe EventSubscribe) error
	ProcessUnsubscribe(eventUnsubscribe EventUnsubscribe) error

	ProcessTimeout(eventTimeout EventTimeout) error
	ProcessIOException(eventIOException EventIOException) error
}

////////////////////Implementation////////////////////////
