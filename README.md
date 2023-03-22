<h1>
bc-explorer: a comprehensive blockchain explorer 
</h1>
bc-explorer is a block explorer for **bestchains** which has three components :

- `viewer`: view formatted blockchain network data with `http apis`,including:
    - `networks`
    - `blocks`
    - `transactions`
- `listener`: listen on blockchain network events and inject formatted data to database(`postgresql`).Also support:
    - `register` a new blockchain network
    - `deregister` a blockchain network
- `observer`: observe network status in `bestchains` platform and automatically register/deregister networks into `listener`


> NOTE: For API authorization & authentication,we will use [kube-rbac-proxy](https://github.com/brancz/kube-rbac-proxy).

![Architecture](./doc/images/arch.png)

### Prerequsities

- [Go1.18]()
- [Postgresql](https://www.postgresql.org/download/)


### Quick start

#### Listener
1. build bc-explorer listener

```
go build -o bin/listener cmd/listener/main.go
```

2. verify `listener`

```
Usage of ./bin/listener:
  -add_dir_header
        If true, adds the file directory to the header of the log messages
  -addr string
        used to listen and serve http requests (default ":9999")
  -alsologtostderr
        log to standard error as well as files (no effect when -logtostderr=true)
  -dsn string
        database conneciton string (default "postgres://bestchains:Passw0rd!@127.0.0.1:5432/bc-explorer?sslmode=disable")
  -injector string
        used to initialize injector (default "pg")
  -log_backtrace_at value
        when logging hits line file:N, emit a stack trace
  -log_dir string
        If non-empty, write log files in this directory (no effect when -logtostderr=true)
  -log_file string
        If non-empty, use this log file (no effect when -logtostderr=true)
  -log_file_max_size uint
        Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
  -logtostderr
        log to standard error instead of files (default true)
  -one_output
        If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true)
  -skip_headers
        If true, avoid header prefixes in the log messages
  -skip_log_headers
        If true, avoid headers when opening log files (no effect when -logtostderr=true)
  -stderrthreshold value
        logs at or above this threshold go to stderr when writing to files and stderr (no effect when -logtostderr=true or -alsologtostderr=false) (default 2)
  -v value
        number for the log level verbosity
  -vmodule value
        comma-separated list of pattern=N settings for file-filtered logging
```

3. start bc-explorer listener

```
./bin/listener -addr localhost:9999 -injector pg -dsn postgres://username:password@127.0.0.1:5432/bc-explorer?sslmode=disable
```

#### Viewer
1. build bc-explorer viewer

```
go build -o bin/viewer cmd/viewer/main.go
```

2. verify `viewer`

```
Usage of ./bin/viewer:
  -add_dir_header
        If true, adds the file directory to the header of the log messages
  -addr string
        used to listen and serve http requests (default ":9998")
  -alsologtostderr
        log to standard error as well as files (no effect when -logtostderr=true)
  -db string
        which database to use, default is pg(postgresql) (default "pg")
  -dsn string
        database conneciton string (default "postgres://bestchains:Passw0rd!@127.0.0.1:5432/bc-explorer?sslmode=disable")
  -log_backtrace_at value
        when logging hits line file:N, emit a stack trace
  -log_dir string
        If non-empty, write log files in this directory (no effect when -logtostderr=true)
  -log_file string
        If non-empty, use this log file (no effect when -logtostderr=true)
  -log_file_max_size uint
        Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
  -logtostderr
        log to standard error instead of files (default true)
  -one_output
        If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true)
  -skip_headers
        If true, avoid header prefixes in the log messages
  -skip_log_headers
        If true, avoid headers when opening log files (no effect when -logtostderr=true)
  -stderrthreshold value
        logs at or above this threshold go to stderr when writing to files and stderr (no effect when -logtostderr=true or -alsologtostderr=false) (default 2)
  -v value
        number for the log level verbosity
  -vmodule value
        comma-separated list of pattern=N settings for file-filtered logging
```

3. start bc-explorer viewer

```bash
# test the service by logging, which will print the request and return a false data
./bin/viewer -v=5 -db=log 

# connect to pg database test service
./bin/viewer -v=5 -dsn='postgres://username:password@127.0.0.1:5432/bc-explorer?sslmode=disable'
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
