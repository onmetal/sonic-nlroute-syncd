# sonic-nlroute-syncd
Subscribes to route changes from netlink and synchronizes them into
the SONiC APPL_DB (0).

## Limitations
Only regular IPv4/IPv6 routes are supported. No support for routes within VRFs. No support for MPLS.