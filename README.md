# gexpose

A net tool for exposing local service to public.

[![Travis](https://travis-ci.com/net-byte/gexpose.svg?branch=main)](https://github.com/net-byte/gexpose)
[![Go Report Card](https://goreportcard.com/badge/github.com/net-byte/gexpose)](https://goreportcard.com/report/github.com/net-byte/gexpose)
![image](https://img.shields.io/badge/License-MIT-orange)
![image](https://img.shields.io/badge/License-Anti--996-red)

# Architecture
<p>
	<img src="https://github.com/net-byte/gexpose/raw/main/assets/gexpose.png" alt="gexpose" width="900">
</p>

# Usage

```
Usage of ./gexpose:
  -server
        server mode
  -k string
        encryption key (default "Xn2r4u7x!A%D*G8")
  -l string
        local address (default ":9000")
  -s string
        server address (default ":8701")
  -p string
        proxy address (default ":8702")
  -e string
        expose address (default ":8703")
  -t int
        dial timeout in seconds (default 30)

```

## Build

```
sh scripts/build.sh
```

## Docker

### Run client
```
docker run  -d --privileged --restart=always --net=host --name gexpose-client netbyte/gexpose -s server-addr:8701 -p server-addr:8702 -l 127.0.0.1:8080
```

### Run server
```
docker run  -d --privileged --restart=always --net=host --name gexpose-server netbyte/gexpose -server
```

# License
[The MIT License (MIT)](https://raw.githubusercontent.com/net-byte/gexpose/main/LICENSE)
