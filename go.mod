module github.com/dubbogo/gost

require (
	github.com/Workiva/go-datastructures v1.0.52
	github.com/apache/dubbo-go v1.5.6
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/dubbogo/go-zookeeper v1.0.3
	github.com/dubbogo/jsonparser v1.0.1
	github.com/jinzhu/copier v0.0.0-20190625015134-976e0346caa8
	github.com/k0kubun/pp v3.0.1+incompatible
	github.com/mattn/go-isatty v0.0.12
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/shirou/gopsutil v3.20.11+incompatible
	github.com/stretchr/testify v1.7.0
	go.uber.org/atomic v1.7.0
	google.golang.org/grpc v1.33.1
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	go.etcd.io/bbolt v1.3.4 => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc v1.33.1 => google.golang.org/grpc v1.26.0
)

go 1.13
