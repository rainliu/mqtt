mqtt
====
<br>
MQTT in Go<br>
<br>
<br>
MQTT Conformance/Interoperability Testing<br>
<br>
https://eclipse.org/paho/clients/testing/<br>
<br>
python3 client_test.py [hostname:port]:<br>
<br>
Traceback (most recent call last):<br>
  File "client_test.py", line 231, in offline_message_queueing_test<br>
    assert len(callback.messages) in [2, 3], callback.messages<br>
AssertionError: []<br>
Offline message queueing test failed<br>
<br>
Redelivery on reconnect test starting<br>
Traceback (most recent call last):<br>
  File "client_test.py", line 309, in redelivery_on_reconnect_test<br>
    assert len(callback2.messages) == 2, "length should be 2: %s" % callback2.messages<br>
AssertionError: length should be 2: []<br>
Redelivery on reconnect test failed<br>
test suite failed<br>
<br>
<br>
hostname localhost port 1883<br>
clean up starting<br>
clean up finished<br>
Basic test starting<br>
Basic test succeeded<br>
Retained message test starting<br>
Retained message test succeeded<br>
Will message test succeeded<br>
Overlapping subscriptions test starting<br>
This server is publishing one message for all matching overlapping subscriptions, not one for each.<br>
Overlapping subscriptions test succeeded<br>
Keepalive test starting<br>
Keepalive test succeeded<br>
Zero length clientid test starting<br>
Zero length clientid test succeeded<br>
Subscribe failure test starting<br>
Subscribe failure test succeeded<br>
$ topics test starting<br>
$ topics test succeeded<br>
test suite succeeded<br>

<br>
$ sdkperf_mqtt.sh -cip=localhost -ptl=T/demo -stl=T/demo -mpq=1 -msq=1 -msa=100
 -mn=10000 -mr=10000000
<br>
TraceOn():<br>
Total Messages transmitted = 10000<br>
Computed publish rate (msg/sec) = 1122.0<br>
-------------------------------------------------
Total Messages received across all subscribers = 10000<br>
Messages received with discard indication = 0<br>
Computed subscriber rate (msg/sec across all subscribers) = 1123<br>

TraceOff():<br>
-------------------------------------------------
Total Messages transmitted = 10000<br>
Computed publish rate (msg/sec) = 7694.0<br>
-------------------------------------------------
Total Messages received across all subscribers = 10000<br>
Messages received with discard indication = 0<br>
Computed subscriber rate (msg/sec across all subscribers) = 7705<br>