package main

import (
	"mqtt"
)

type mqtts_listener struct {
	provider mqtt.Provider
}

func newListener(provider mqtt.Provider) *mqtts_listener {
	return &mqtts_listener{provider}
}
func (this *mqtts_listener) ProcessSessionCreated(eventSessionCreated mqtt.EventSessionCreated) {
	println("Session Created with reason: ", eventSessionCreated.GetReason())
}
func (this *mqtts_listener) ProcessSessionTerminated(eventSessionTerminated mqtt.EventSessionTerminated) {
	println("Session Terminated with reason: ", eventSessionTerminated.GetReason())
}

func (this *mqtts_listener) ProcessConnect(eventConnect mqtt.EventConnect) {
	println("Received CONNECT")
	serverSession := eventConnect.GetSession().(mqtt.ServerSession)
	serverSession.Respond(false, mqtt.CONNACK_RETURNCODE_ACCEPTED)
}
func (this *mqtts_listener) ProcessPublish(eventPublish mqtt.EventPublish) {

}
func (this *mqtts_listener) ProcessSubscribe(eventSubscribe mqtt.EventSubscribe) {

}
func (this *mqtts_listener) ProcessUnsubscribe(eventUnsubscribe mqtt.EventUnsubscribe) {

}

func (this *mqtts_listener) ProcessTimeout(eventTimeout mqtt.EventTimeout) {

}
func (this *mqtts_listener) ProcessIOException(eventIOException mqtt.EventIOException) {
	println("Received IOException")
}
