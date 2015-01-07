package mqtt

////////////////////Interface//////////////////////////////

type Listener interface {
	ProcessSessionCreated(sessionCreatedEvent SessionCreatedEvent)
	ProcessSessionTerminated(sessionTerminatedEvent SessionTerminatedEvent)

	ProcessPublish(publishEvent PublishEvent)
	ProcessSubscribe(subscribeEvent SubscribeEvent)
	ProcessUnsubscribe(unsubscribeEvent UnsubscribeEvent)

	ProcessTimeout(timeoutEvent TimeoutEvent)
	ProcessIOException(ioExceptionEvent IOExceptionEvent)
}

////////////////////Implementation////////////////////////
