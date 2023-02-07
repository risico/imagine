# imagine
![poors man ai generated project image](https://i.imgur.com/xJPvrPI.png)

Imagine is a image serving service that can act as its own server or be embeded as a package directly into your application. This package is built upon the bimg go package that itself is build upon
the libvips code. It is a very fast image encoding/decoding library.

## Installation

```
go get github.com/risico/imagine/cmd
```

```
imagine start
```

or import it in your own application:

```
package main

import (
    "github.com/risico/imagine"
    "github.com/risico/imagine/stores/redis"
    "github.com/risico/imagine/adaptor"
)

func main() {
    i, _ := imagine.New(imagine.Params{
        Storage: stores.NewLocalStorage(stores.LocalStorageParams{}),
        Cache: stores.NewRedisCache(stores.RedisCacheParams{
            TTL: 10 * time.Hour,
        }),
    })

    // given a router or something
    ginAdaptor := adaptors.NewGinAdaptor(NetAdaptorParams{Imagine: i})
    router.RouterGroup(ginAdaptor)

    // or you can use the data directly although you would need to implement the routers yourself

    params := i.ParamsFromQueryString(u.Query())
    img, err := i.Get(url.Path, params)
    // show the image yourself
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
