# Bestchains Explorer

[![codecov](https://codecov.io/gh/bestchains/bc-explorer/branch/main/graph/badge.svg?token=6W2QTWHQY1)](https://codecov.io/gh/bestchains/bc-explorer)

bc-explorer is a block explorer for **bestchains** which has three components :

- `viewer`: view formatted blockchain network data with `http apis`,including:
  - `networks`
  - `blocks`
  - `transactions`
- `listener`: listen on blockchain network events and inject formatted data to database(`postgresql`).Also support:
  - `register` a new blockchain network
  - `deregister` a blockchain network
- `observer`: observe network status in `bestchains` platform and automatically register/deregister networks into `listener`
- `client`: fabric test client to help generate contract calls

> NOTE: For API authorization & authentication,we allow tree ways
>
> - `none`: no authorization & authentication
> - `oidc`: oidc authorization & authentication
> - `kubernetes`: kubernetes authorization & authentication

![Architecture](./doc/images/arch.png)

## Usage

### Prerequsities

- [Go1.20](https://go.dev/doc/install)
- [Postgresql](https://www.postgresql.org/download/)

### Build Image

```bash
# output: hyperledgerk8s/bc-explorer:v0.1.0
WHAT=bc-explorer GOOS=linux GOARCH=amd64 make image
```

### Quick start

#### Listener

1. build bc-explorer listener

```shell
go build -o bin/listener cmd/listener/main.go
```

2. verify `listener`

```shell
Usage of ./bin/listener:
  -add_dir_header
     If true, adds the file directory to the header of the log messages
  -addr string
     used to listen and serve http requests (default ":9999")
  -alsologtostderr
     log to standard error as well as files
  -auth string
     user authentication method, none, oidc or kubernetes (default "none")
  -dsn string
     database connection string (default "postgres://bestchains:Passw0rd!@127.0.0.1:5432/bc-explorer?sslmode=disable")
  -injector string
     used to initialize injector (default "pg")
  -kubeconfig string
     Paths to a kubeconfig. Only required if out-of-cluster.
  -log_backtrace_at value
     when logging hits line file:N, emit a stack trace
  -log_dir string
     If non-empty, write log files in this directory
  -log_file string
     If non-empty, use this log file
  -log_file_max_size uint
     Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
  -logtostderr
     log to standard error instead of files (default true)
  -one_output
     If true, only write logs to their native severity level (vs also writing to each lower severity level)
  -skip_headers
     If true, avoid header prefixes in the log messages
  -skip_log_headers
     If true, avoid headers when opening log files
  -stderrthreshold value
     logs at or above this threshold go to stderr (default 2)
  -v value
     number for the log level verbosity
  -vmodule value
     comma-separated list of pattern=N settings for file-filtered logging
```

3. start bc-explorer listener

```shell
./bin/listener -addr localhost:9999 -injector pg -dsn postgres://username:password@127.0.0.1:5432/bc-explorer?sslmode=disable
```

#### Viewer

1. build bc-explorer viewer

```shell
go build -o bin/viewer cmd/viewer/main.go
```

2. verify `viewer`

```shell
Usage of ./bin/viewer:
  -add_dir_header
     If true, adds the file directory to the header of the log messages
  -addr string
     used to listen and serve http requests (default ":9998")
  -alsologtostderr
     log to standard error as well as files
  -auth string
     user authentication method, none, oidc or kubernetes (default "none")
  -db string
     which database to use, default is pg(postgresql) (default "pg")
  -dsn string
     database connection string (default "postgres://bestchains:Passw0rd!@127.0.0.1:5432/bc-explorer?sslmode=disable")
  -kubeconfig string
     Paths to a kubeconfig. Only required if out-of-cluster.
  -log_backtrace_at value
     when logging hits line file:N, emit a stack trace
  -log_dir string
     If non-empty, write log files in this directory
  -log_file string
     If non-empty, use this log file
  -log_file_max_size uint
     Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
  -logtostderr
     log to standard error instead of files (default true)
  -one_output
     If true, only write logs to their native severity level (vs also writing to each lower severity level)
  -skip_headers
     If true, avoid header prefixes in the log messages
  -skip_log_headers
     If true, avoid headers when opening log files
  -stderrthreshold value
     logs at or above this threshold go to stderr (default 2)
  -v value
     number for the log level verbosity
  -vmodule value
     comma-separated list of pattern=N settings for file-filtered logging
```

3. start bc-explorer viewer

```shell
# test the service by logging, which will print the request and return a false data
./bin/viewer -v=5 -db=log 

# connect to pg database test service
./bin/viewer -v=5 -dsn='postgres://username:password@127.0.0.1:5432/bc-explorer?sslmode=disable'
```

#### Client

1. build bc-explorer client

```shell
go build -o bin/client cmd/client/*
```

2. verify `client`

```shell
Usage of ./bin/client:
  -args value
        a list of arguments for contract call
  -contract string
        contract name (default "samplecc")
  -method string
        contract method (default "PutValue")
  -profile string
        profile to connect with blockchain network (default "./network.json")
```

3. example for `client`

For a contract [`samplecc`](https://github.com/bestchains/fabric-builder-k8s/blob/main/samples/go-contract/main.go), we use `client` to call `PutValue`

```shell
./bin/client -profile ./test/sample_fabric_network.json -contract samplecc -method PutValue -args platform -args bestchains
```

After this contract call, a transaction will be injected to `bc-explorer` with network id `blkexp_blkexp6`. A pair of `Key-Value`(`{"platform":"bestchains"}`) will be stored into blockchain statedb.

Contract call can be verified by  `GetValue`

```shell
./bin/client -profile ./test/sample_fabric_network.json -contract samplecc -method GetValue -args platform
```

Output:

```shell
I0324 10:24:01.523211   21170 main.go:71] Result: bestchains
```

#### observer

1. build bc-explorer observer

```shell
go build -o bin/observer cmd/observer/main.go
```

2. verify `observer`

```shell
Usage of ./bin/observer:
  -add_dir_header
     If true, adds the file directory to the header of the log messages
  -alsologtostderr
     log to standard error as well as files
  -auth string
     user authentication method, none, oidc or kubernetes (default "none")
  -host string
     the host of listener (default "http://localhost:9999")
  -kubeconfig string
     Paths to a kubeconfig. Only required if out-of-cluster.
  -log_backtrace_at value
     when logging hits line file:N, emit a stack trace
  -log_dir string
     If non-empty, write log files in this directory
  -log_file string
     If non-empty, use this log file
  -log_file_max_size uint
     Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
  -logtostderr
     log to standard error instead of files (default true)
  -one_output
     If true, only write logs to their native severity level (vs also writing to each lower severity level)
  -operator-namespace string
     the ns of fabric-operator (default "baas-system")
  -skip_headers
     If true, avoid header prefixes in the log messages
  -skip_log_headers
     If true, avoid headers when opening log files
  -stderrthreshold value
     logs at or above this threshold go to stderr (default 2)
  -v value
     number for the log level verbosity
  -vmodule value
     comma-separated list of pattern=N settings for file-filtered logging
```

3. start bc-explorer observer

```shell
# test the service by logging
./bin/observer -v=5

# out of cluster
./bin/observer -v=5 --host host_of_listener --kubeconfig ~/.kube/config
```

## Development

### Models

[See the documentation](./doc/models.md)

### APIs

- `Viewer` APIs : [See the documentation](./doc/viewer_apis.md)
- `Listener` APIs : [See the documentation](./doc/listener_api.md)

## Contribute to bc-explorer

If you want to contribute to bc-explorer,refer to [contribute guide](./CONTRIBUTING.md)

## Support

If you need support, start with the troubleshooting guide, or create github [issues](https://github.com/bestchains/bc-explorer/issues/new)
