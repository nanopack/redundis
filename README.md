[![redundis logo](http://nano-assets.gopagoda.io/readme-headers/redundis.png)](http://nanobox.io/open-source#redundis)
[![Build Status](https://travis-ci.org/nanopack/redundis.svg)](https://travis-ci.org/nanopack/redundis)

# Redundis

Redis high-availability cluster using Sentinel to transparently proxy connections to the active primary member.

Redundis is a smart sentinel aware proxy for redis that allows redis clients to not care about failover of the redis master node.

Connections are automatically forwarded to the master redis node, and when the master node fails over, clients are disconnected and reconnected to the new master node.

Redundis was written in luvit because of its near c speed, and because it uses a very small amount of memory.


## Status

Complete/Stable

# Config File

The default location for the config file is `/opt/local/etc/redundis/redundis.conf`. This can be changed by passing the new location in as the first parameter to the redundis command: `redundis.lua ./path/to/file.config`.

### Config Params

#### sentinel_ip
- default '127.0.0.1'
- ip for sentinel connection

#### sentinel_port
- default 26379
- port for sentinel connection

#### listen_ip
- default '127.0.0.1'
- ip for redundis to listen for clients

#### listen_port
- default 6379
- port for redundis to listen for clients

#### monitor_name
- default 'test'
- name of server to proxy for, this is defined in the sentinal.conf file.

#### not_ready_timeout
- default 5000
- retry timeout for sentinels not being ready

#### sentinel_poll_timeout
- default 1000
- polling interval to check for a new master

#### master_wait_timeout
- default 1000
- how long to wait before checking if the new master is ready


# Limitations

Currently only connecting to one sentinel is supported. It could be extended in the future to connect to a different sentinel incase of sentinel failure, but right now this is not needed.

[![redundis logo](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
