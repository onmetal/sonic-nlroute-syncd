module github.com/onmetal/sonic-nlroute-syncd

go 1.16

require (
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/onsi/ginkgo v1.16.1 // indirect
	github.com/onsi/gomega v1.11.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/vishvananda/netlink v1.1.0
	golang.org/x/sys v0.0.0-20210603081109-ebe580a85c40 // indirect
)

replace github.com/vishvananda/netlink => github.com/taktv6/netlink v1.1.1-0.20210519175051-ac6e361bed8f
