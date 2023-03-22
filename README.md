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
./bin/listener -h
Usage of ./bin/listener:
  -addr string
        used to listen and serve http requests (default ":9999")
  -dsn string
        database conneciton string (default "postgres://bestchains:Passw0rd!@127.0.0.1:5432/bc-explorer?sslmode=disable")
  -injector string
        used to initialize injector (default "pg")
```

3. start bc-explorer listener

```
./bin/listener -addr localhost:9999 -injector pg -dsn postgres://bestchains:Passw0rd!@127.0.0.1:5432/bc-explorer?sslmode=disable
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
