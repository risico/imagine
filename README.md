# imagine

Imagine is a image serving service that can act as its own server or be embeded as a package directly into your application.

## Installation

```
go get github.com/risico/imagine/v1
```

```
imagine start
```

or import it in your own application:

```
package main

import github.com/risico/imagine/v1

func main() {
    i : imagine.New(imagine.Params{
        Storage: storage.NewLocalStorage(cache.LocalStorageParams{}),
        Cache: cache.NewRedisCache(cache.RedisCacheParams{}),
    })

    err := i.RegisterRoutes(router)
    if err != nil { }
}

```

## Standalone Server Config

```
server:
    hostname: localhost
    port: 1234
    ssl: true -- automatically create a SSL cert
authentication:
    enabled: true
    password: w
storage:
    mode: local | s3
cache:
    mode: Memory | Local | Redis
```
