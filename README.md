# Overview

Simple toy server to experiment with a variety of go features. The ultimate goal is to see how much trouble I can get it into by spawning a large number of clients with a short timeout, causing transactions to be canceled.

Once started, the server listens for tcp connections on port 10000. Commands should be newline terminated.

# Building

go build -o toyserver

# Prerequisites

This expects a mysql compatible database to be listening on 3307.
E.g.

```
docker pull mariadb
docker run --name toyserver -e MYSQL_ROOT_PASSWORD=<your password> -p 3307:3306 mariadb:latest
```

-Note this will leave the container hanging around, use --rm to remove

# Usage

DBUSER=<user> DBPASS=<pass> ./toyserver

telnet localhost 10000
