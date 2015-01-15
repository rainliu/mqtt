package main

import (
	"log"
	"mqtt"
)

type mqtts_listener struct {
	provider mqtt.Provider
}

func newListener(provider mqtt.Provider) *mqtts_listener {
	return &mqtts_listener{provider: provider}
}
func (this *mqtts_listener) ProcessConnect(eventConnect mqtt.EventConnect) {
	log.Println("Received CONNECT")
	serverSession := eventConnect.GetSession().(mqtt.ServerSession)

	pktconnack := mqtt.NewPacketConnack()
	pktconnack.SetSPFlag(false)
	pktconnack.SetReturnCode(mqtt.CONNACK_RETURNCODE_ACCEPTED)

	serverSession.Acknowledge(pktconnack)
}
func (this *mqtts_listener) ProcessPublish(eventPublish mqtt.EventPublish) {
	log.Printf("Received Publish with DUP %v QoS %v RETAIN %v Topic: %v Content: %v\n",
		eventPublish.GetMessage().GetDup(),
		eventPublish.GetMessage().GetQos(),
		eventPublish.GetMessage().GetRetain(),
		eventPublish.GetMessage().GetTopic(),
		eventPublish.GetMessage().GetContent())

	this.provider.Forward(eventPublish.GetMessage())
}
func (this *mqtts_listener) ProcessSubscribe(eventSubscribe mqtt.EventSubscribe) {
	log.Printf("Received SUBSCRIBE with %v", eventSubscribe.GetSubscribeTopics())
	serverSession := eventSubscribe.GetSession().(mqtt.ServerSession)

	pktsuback := mqtt.NewPacketSuback()
	pktsuback.SetPacketId(eventSubscribe.GetPacketId())
	qos := eventSubscribe.GetQoSs()
	retCodes := make([]byte, len(qos))
	for i := 0; i < len(qos); i++ {
		retCodes[i] = byte(qos[i])
	}
	pktsuback.SetReturnCodes(retCodes)

	serverSession.Acknowledge(pktsuback)
}
func (this *mqtts_listener) ProcessUnsubscribe(eventUnsubscribe mqtt.EventUnsubscribe) {
	log.Printf("Received UNSUBSCRIBE with %v", eventUnsubscribe.GetUnsubscribeTopics())
}

func (this *mqtts_listener) ProcessTimeout(eventTimeout mqtt.EventTimeout) {
	log.Println("Received Timeout")
}
func (this *mqtts_listener) ProcessIOException(eventIOException mqtt.EventIOException) {
	log.Println("Received IOException")
}
func (this *mqtts_listener) ProcessSessionTerminated(eventSessionTerminated mqtt.EventSessionTerminated) {
	log.Println("Session Terminated with Reason: ", eventSessionTerminated.GetReason())
}
