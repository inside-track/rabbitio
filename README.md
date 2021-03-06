<p align="center"><img src="https://user-images.githubusercontent.com/3073246/38677615-11f18176-3e5f-11e8-935f-b1b824e53e92.png" width="300"></p>

# RabbitIO
[![Build Status](https://travis-ci.org/meltwater/rabbitio.svg?branch=master)](https://travis-ci.org/meltwater/rabbitio)

RabbitIO is a tool to backup and restore RabbitMQ messages. 

To learn more about common problems with queue management and our reasoning and design principles when building RabbitIO, please read this [blog post](https://underthehood.meltwater.com/blog/2018/08/03/rabbitio-a-tool-to-backup-and-restore-messages-from-rabbitmq/).

### Typical use case

A typical use case is to consume all messages from a RabbitMQ deadletter queue to your local dev machine, do some analysis of the messages, and possibly push (some of) them back to an exchange for reprocessing.

## Installing

#### Download binary

Pick your binary from [Releases](https://github.com/meltwater/rabbitio/releases) and download, in addition you'll need to set the binary to be executable.

Example with `linux-amd64` and version `v0.5.4`:

```bash
wget https://github.com/meltwater/rabbitio/releases/download/v0.5.4/rabbitio-v0.5.4-linux-amd64 -O rabbitio
chmod 755 rabbitio
```

#### Using `go get`

```bash
go get -u github.com/meltwater/rabbitio
```

## How to use RabbitIO

After installing rabbitio, you can quickly test out `rabbitio` by using [docker-compose](https://docs.docker.com/compose/install/).

Included is a docker-compose file to set up local rabbitmq.

```bash
cd $GOPATH/src/github.com/meltwater
docker-compose up -d
```

Go to your now running [local rabbit](http://localhost:15672) and create example exchange `rabbitio-exchange` and queue `rabbitio-queue`. Then bind the queue to the exchange.

#### Publish your first message

```bash
echo "My first message" > message # write a message into a file
tar cfz message.tgz message # create a tarball containing this message
rabbitio in -e rabbitio-exchange -q rabbitio-queue -f message.tgz
```

This will publish your first message into `rabbitio-exchange` and you'll see your message in the queue `rabbitio-queue`

#### Consume your first message

```bash
$ mkdir data
$ rabbitio out -e rabbitio-exchange -q rabbitio-queue -d data/
2018/03/15 15:37:35 RabbitMQ connected: amqp://guest:guest@localhost:5672/
2018/03/15 15:37:35 Bind to Exchange: "rabbitio-exchange" and Queue: "rabbitio-queue", Messaging waiting: 1
^C Interruption, saving last memory bits..
2018/03/15 15:37:45 Wrote 203 bytes to data/1_messages_1.tgz
2018/03/15 15:37:45 tarball writer closing
```

We interrupt when the queue is empty by directly using a combination of `CTRL + C` once. This will save the last bits and ack the message.

## Detailed Usage

```
$ rabbitio
Rabbit IO will help backup and restore your messages in RabbitMQ

Usage:
  rabbitio [command]

Available Commands:
  help        Help about any command
  in          Publishes documents from tarballs into RabbitMQ exchange
  out         Consumes data out from RabbitMQ and stores to tarballs
  version     Prints the version of Rabbit IO

Flags:
  -e, --exchange string     Exchange to connect to
  -h, --help                help for rabbitio
  -p, --prefetch int        Prefetch for batches (default 100)
  -q, --queue string        Queue to connect to
  -r, --routingkey string   Routing Key, if specified will override tarball routing key configuration (default "#")
  -t, --tag string          AMQP Client Tag (default "Rabbit IO Connector ")
  -u, --uri string          AMQP URI, uri to for instance RabbitMQ (default "amqp://guest:guest@localhost:5672/")

Use "rabbitio [command] --help" for more information about a command.
```

### AMQP Headers and Routing Key

When you read messages from a queue, the headers as well as the routing key will be saved as metadata in the tarballs, utilizing what in tar is called PAX Records. This is helpful if you one day want to replay the data back into the original queue, while keeping the attributes that belong to the message.

|    Header Format   |                AMQP Headers               |   Tar PAX Records  |
|:------------------:|:-----------------------------------------:|:------------------:|
| Format Translation | map[String] Bool, Integer, String, Float  | map[String] String |
| Body Type          | Bytes                                     | Bytes              |

The tar metadata can be accessed using [pax](https://linux.die.net/man/1/pax):

```bash
pax -r -zf 1_message_100.tgz
```

This will output the messages and in addition a `PaxHeaders.0` directory containing identical filenames as the messages, enabling access of the metadata. Right now there is no way to selectively publish only certain messages containing a header value, this is a planned feature.

## Contributing

If you plan to work on `rabbitio` you will need [Golang](https://golang.org/dl/). PRs are welcome as well as implementation discussions.

**Clone `rabbitio`**
```bash
mkdir -p $GOPATH/src/github.com/meltwater
cd $GOPATH/src/github.com/meltwater
git clone git@github.com:meltwater/rabbitio.git
```

#### Building

```bash
cd rabbitio
make && make build
```

## Maintainers

For any bug reports or change requests, please create a GitHub issue or submit a PR.

Also feel free to drop a line to the maintainers:

- Joel ([@vorce](https://github.com/vorce))
- Stian ([@stiangrindvoll](https://github.com/stiangrindvoll))
