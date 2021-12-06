/*

1. only join() and stabilize() could update your node's pred n succ ?? 

2. why only transfer data when "notified" i am your new predecessor 
but not during join() or "stabilize" i am your new successor  ??

3. why no networking call + thread + channel pattern ??
A: implemented in rpc lib, could config to same/diff packet send many times, many rpc wait for same replies









*/