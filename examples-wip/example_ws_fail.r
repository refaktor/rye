new-server ":8080"
 |handle "/" fn { w r } { write w "Hello from Rye function!" }  
 |handle-ws "/ping" fn { s ctx } { print read s ctx print "before fail" failure "failure in handler" |print print "after fail" }   
 |serve
