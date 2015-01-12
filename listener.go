package mqtt

////////////////////Interface//////////////////////////////

type Listener interface {
	ProcessSessionCreated(eventSessionCreated EventSessionCreated)
	ProcessSessionTerminated(eventSessionTerminated EventSessionTerminated)

	ProcessConnect(eventConnect EventConnect)
	ProcessPublish(eventPublish EventPublish)
	ProcessSubscribe(eventSubscribe EventSubscribe)
	ProcessUnsubscribe(eventUnsubscribe EventUnsubscribe)

	ProcessTimeout(eventTimeout EventTimeout)
	ProcessIOException(eventIOException EventIOException)
}

////////////////////Implementation////////////////////////
