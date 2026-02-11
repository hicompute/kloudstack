## Default Internal Network
on ovn central node(s):

```bash
ovn-nbctl ls-add public
ovn-nbctl lsp-add public ln-public
ovn-nbctl lsp-set-addresses ln-public unknown
ovn-nbctl lsp-set-type ln-public localnet
ovn-nbctl lsp-set-options ln-public network_name=public
```

## External Network

on every worker should create an external bridge:

```bash
ovs-vsctl add-br br-ext
ovs-vsctl add-port br-ext provider1
ovs-vsctl set open . external-ids:ovn-bridge-mappings="public:br-ext"
```
