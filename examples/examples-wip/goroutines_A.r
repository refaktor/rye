event: new-channel 24
hub: fn { e b } { print "HUB started" forever { print "HUB waiting" read e |pass { prn "HUB received:" } |print } }
go-with event ?hub
sleep 2000
send event "does this work?"
sleep 2000
send event "again?"
sleep 2000
print "I and going away"