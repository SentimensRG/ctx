# ctx

Composable utilities for Go contexts.

[![Godoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/SentimensRG/ctx)
[![Go Report Card](https://goreportcard.com/badge/github.com/SentimensRG/ctx?style=flat-square)](https://goreportcard.com/report/github.com/SentimensRG/ctx)

## Installation

```bash
go get -u github.com/SentimensRG/ctx
```

## Features

### Ctx

#### General context / done-channel utilities

The `ctx` package provides utilites for working with data structures satisfying
the `ctx.Doner` interface, most notably `context.Context`:

```go
type Doner interface {
    Done() <-chan struct{}
}
```

The functions in `ctx` are appropriate for operations that do not preserve the
values in a context, e.g.: joining several contexts together.

### SigCtx

#### Go contexts for graceful shutdown

The `sigctx` package provides a context that terminates when it receives a
SIGINT or SIGTERM.  This provides a convenient mechanism for triggering
graceful application shutdown.

#### Usage

`sigctx.New` returns a `ctx.C`, which implements the ubiquitous `ctx.Doner`
interface.  It fires when either SIGINT or SIGTERM is caught.

```go
import (
    "log"
    "github.com/SentimensRG/ctx/sigctx"
)

func main() {
    ctx := sigctx.New()  // returns a regular context.Context

    <-ctx.Done()  // will unblock on SIGINT and SIGTERM
    log.Println("exiting.")
}
```

`sigctx.Tick` can be used to react to streams of signals.  For example, you can
implement a graceful shutdown attempt, followed by a forced shutdown.

```go
import (
    "log"
    "github.com/SentimensRG/ctx/sigctx"
    "github.com/SentimensRG/ctx"
)

func main() {
    t := sigctx.Tick()
    d, cancel := ctx.WithCancel(ctx.Background())

    go func() {
        defer cancel()

        go func() {
            // business logic goes here
        }()

        <-t
        log.Println("attempting graceful shutdown - press Ctrl + c again to force quit")
        go func() {
            defer cancel()
            // cleanup logic goes here
        }()

        <-t
        log.Println("forcing close")
    }()

    <-d.Done()
}

```

### RefCtx

#### Contextualized reference counting

The `refctx` package provides a context whose lifetime is bound to a `RefCtr`
instance.  The `RefCtr` keeps track of how many references to a context are
held and cancels this context when the refcount reaches zero.

#### Usage

`refctx` works similarly to `sync.WaitGroup`.

```go
package main

import "github.com/SentimensRG/ctx/refctx"

func main() {
    ctx, ctr := refctx.WithRefCount(context.Background())

    for i := 0; i < 5; i++ {
        ctr.Incr()
        go func() {
            defer ctr.Decr()

            time.Sleep(time.Second * i)
        }()
    }

    <-ctx.Done()  // fires when refcount falls back to zero
}

```

A common use-case for `refctx` is to manage timeouts.  Consider the following
example using `github.com/gorilla/websocket`.

```go
import (
    "time"

    "github.com/SentimensRG/ctx/refctx"
    "github.com/SentimensRG/ctx/sigctx"

    "github.com/gorilla/websocket"
)

const (
    pingDeadline = time.Second * 1
    pongDeadline = pingDeadline * 2
)

func main() {

    conn := openWebsocketConnection()

    ctx, ctr := refctx.WithRefCount(sigctx.New())  // good place for sigctx
    rc.Incr()  // start with one refcount

    go func() {
        for range time.NewTicker(heartbeatInterval).C {
            select {
            case <-c.Done():
                // c.Done fires either when the process receives an OS signal, or
                // when the peer took too long to respond to a ping.
                return
            default:
                deadline := time.Now().Add(pingDeadline)
                _ = conn.WriteControl(websocket.PingMessage, nil, deadline)
                go func() {
                    <-time.After(pongDeadline)
                    rc.Decr()
                }()
            }
        }
    }()

    conn.SetPongHandler(func(_ string) (_ error) {
        rc.Incr()
        return
    })

    businessLogic(c, conn)
    <-c.Done()
}
```

## RFC

If you find this useful please let me know:  <l.thibault@sentimens.com>

Seriously, even if you just used it in your weekend project, I'd like to hear
about it :)

## License

The MIT License

Copyright (c) 2017 Sentimens Research Group, LLC

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
