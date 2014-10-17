# redis-proxy 

Redis-Proxy is a smart sentinal aware proxy for redis that allows redis clients to not care about failover of the redis master node. 

Connections are automatically forwarded to the master redis node, and when the master node fails over, clients are disconnected and reconnected to the new master node.