mqtt
====

MQTT in Go


MQTT Conformance/Interoperability Testing

https://eclipse.org/paho/clients/testing/

python3 client_test.py [hostname:port]:

Traceback (most recent call last):
  File "client_test.py", line 231, in offline_message_queueing_test
    assert len(callback.messages) in [2, 3], callback.messages
AssertionError: []
Offline message queueing test failed

Redelivery on reconnect test starting
Traceback (most recent call last):
  File "client_test.py", line 309, in redelivery_on_reconnect_test
    assert len(callback2.messages) == 2, "length should be 2: %s" % callback2.messages
AssertionError: length should be 2: []
Redelivery on reconnect test failed
test suite failed


hostname localhost port 1883
clean up starting
clean up finished
Basic test starting
Basic test succeeded
Retained message test starting
Retained message test succeeded
Will message test succeeded
Overlapping subscriptions test starting
This server is publishing one message for all matching overlapping subscriptions, not one for each.
Overlapping subscriptions test succeeded
Keepalive test starting
Keepalive test succeeded
Zero length clientid test starting
Zero length clientid test succeeded
Subscribe failure test starting
Subscribe failure test succeeded
$ topics test starting
$ topics test succeeded
test suite succeeded
