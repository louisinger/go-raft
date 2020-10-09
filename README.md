# go-raft

> A Raft algorithm implementation using Golang

Project has an educational aim to learn Golang and how a distributed system works (by implementing RAFT). Raft is a consensus algorithm easy to understand and easy to implement: [more about RAFT](https://raft.github.io/).

The goal of my `go-raft` implementation is to _persist a key-value store accross a set of nodes_.

There is two executables:

- The `goraft-daemon`: the node's implementation of the RAFT algorithm.
- The `goraft-cli`: a command line interface useful to test and communicate with goraft daemons.

**This implementation is not ready for production !**

## Installation

Use the [build script](scripts/build.sh) to compile the sources.

```bash
chmod +x ./scripts/build.sh
./scripts/build.sh
```

## Usage

### Launch a node

_The following command line will launch a node on the local machine serving the port 10001 and will try to contact 2 others nodes: localhost:10002 and localhost:10003_

```bash
./build/goraft-daemon/goraft-daemon 10001 localhost:10002 localhost:10003
```

In this case you need to launch two other process to run the network locally:

```bash
./build/goraft-daemon/goraft-daemon 10002 localhost:10001 localhost:10003
```

```bash
./build/goraft-daemon/goraft-daemon 10003 localhost:10002 localhost:10001
```

I advice you to start nodes in foreground in three different terminals. Thus, you'll be able to read all the logs printed by each node.

### Use the CLI as client

_The following command line will launch a goraft-cli instance, you must specify the node's path as a first parameter._

```bash
./build/goraft-cli/goraft-cli localhost:10002
```

_The goraft-cli serves a simplified interface for testing purposes, see [goraft-cli/main.go](cmd/goraft-cli/main.go)._

Here is a list of the main goraft-cli commands:

- The command **info state** will print the current state of the node (as JSON string).
- The command **put** _key_ _value_ will send new data to the network (for instance **put numberOfStudents 1000** will insert the JSON _{ "value": 1000 }_ at the key _numberOfStudents_).

## Development setup

Install go: https://golang.org/doc/install

## Release History

- 0.0.1 (october 9th 2020)
  - Work in progress

## Meta

Louis SINGER – [@TheSingerLouis](https://twitter.com/dbader_org) – louis@vulpem.com

Distributed under the UNLICENSE license. See `LICENSE` for more information.

## Contributing

1. Fork it (<https://github.com/yourname/yourproject/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request
