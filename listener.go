package mqtt

////////////////////Interface//////////////////////////////

type Listener interface {
	ProcessPublish(publishEvent PublishEvent)
	ProcessSubscribe(subscribeEvent SubscribeEvent)
	ProcessUnsubscribe(unsubscribeEvent UnsubscribeEvent)

	ProcessTimeout(timeoutEvent TimeoutEvent)
	ProcessSessionTerminated(sessionTerminatedEvent SessionTerminatedEvent)
	ProcessIOException(ioExceptionEvent IOExceptionEvent)
}

////////////////////Implementation////////////////////////
