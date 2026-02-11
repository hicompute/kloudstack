# test ovn cluster

```bash
ovs-appctl -t /var/run/ovn/ovnnb_db.ctl cluster/status OVN_Northbound
ovs-appctl -t /var/run/ovn/ovnsb_db.ctl cluster/status OVN_Southbound

ovn-nbctl --db=tcp:192.168.12.177:6641,tcp:192.168.12.178:6641,tcp:192.168.12.179:6641 show
```
