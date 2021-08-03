# sonic-nlroute-syncd
Subscribes to route changes (IPv4/IPv6 unicast) from netlink and synchronizes them into
the SONiC APPL_DB (0). This allows you to run routing daemons other than FRR on SONiC (e.g. BIRD).

## Limitations
Only regular IPv4/IPv6 routes are supported. No support for routes within VRFs. No support for MPLS.

## Building
```
go build
````

## Installation
Make sure you're using an out of band connection when performing the installation.

### Disabling FRR
First you have to make sure SONiCs BGP container is masked.
```
systemctl mask bgp
```

### Installing sonic-nlroute-syncd
Copy the sonic-nlroute-syncd binary to /usr/local/bin.
```
cp sonic-nlroute-syncd /usr/local/bin
```
Place the sonic-nlroute-syncd.service file at /etc/systemd/system/sonic-nlroute-syncd.service.
```
cp sonic-nlroute-syncd.service /etc/systemd/system/sonic-nlroute-syncd.service
```
Reload systend and enable the service:
```
systemctl daemon-reload
systemctl enable sonic-nlroute-syncd
```
Reboot your SONiC switch. Install BIRD and enjoy.