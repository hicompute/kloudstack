## Provider network

and interface (or a vlan or so on) on each worker node must configure as provider network:
netplan:

```yaml
ens33:
    dhcp4: no
    dhcp6: no
    match:
        macaddress: 00:0c:29:91:56:29
    set-name: provider1
```

Better to use jumbo frames (MTU 9000).
