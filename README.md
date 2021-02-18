# gost

[![Build Status](https://travis-ci.org/dubbogo/gost.png?branch=master)](https://travis-ci.org/dubbogo/gost)
[![codecov](https://codecov.io/gh/dubbogo/gost/branch/master/graph/badge.svg)](https://codecov.io/gh/dubbogo/gost)
[![GoDoc](https://godoc.org/github.com/dubbogo/gost?status.svg)](https://godoc.org/github.com/dubbogo/gost)
[![Go Report Card](https://goreportcard.com/badge/github.com/dubbogo/gost)](https://goreportcard.com/report/github.com/dubbogo/gost)
![license](https://img.shields.io/badge/license-Apache--2.0-green.svg)

A go sdk for [Apache Dubbo-go](https://github.com/apache/dubbo-go).

## bytes

* BytesBufferPool
> bytes.Buffer pool

* SlicePool
> slice pool

## container

* queue
> Queue

* set
> HashSet

## log

> output log with color and provides pretty format string

## math

* Decimal

## net

* GetLocalIP() (string, error)
* IsSameAddr(addr1, addr2 net.Addr) bool
* ListenOnTCPRandomPort(ip string) (*net.TCPListener, error) 
* ListenOnUDPRandomPort(ip string) (*net.UDPConn, error)

## page
> Page for pagination. It contains the most common functions like offset, pagesize.

## runtime

* GoSafely 
> Using `go` in a safe way.

* GoUnterminated
> Run a goroutine in a safe way whose task is long live as the whole process life time.

## runtime

* GoSafely 
> Using `go` in a safe way.
* GoUnterminated
> Run a goroutine in a safe way whose task is long live as the whole process life time.

## sync

* TaskPool

## strings

* IsNil
> check a var is nil or not.

## time
> Timer optimization through time-wheel.
