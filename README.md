# Namesilo DNS Operator

Kubernetes operator to watch Ingress objects, and update Namesilo DNS records.

Can be executed "just once" using `update`, or can be run as an operator with `watch`.

Usage:

```
nsdns (update|watch) --domain <domain.name> --ingress-class <some-class>
```
