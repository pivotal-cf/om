## Targeting a director/credhub
The `ssh-private-key` argument is required
when using a local bosh cli without an ssh tunnel.
Otherwise it is optional.

```
eval "$(om bosh-env --ssh-private-key=$KEY_FILE)"
```

## Untargeting a director/credhub
In order to un-target the director/credhub,
the following environment variables need to be unset:

```
unset BOSH_CLIENT
unset BOSH_CLIENT_SECRET
unset BOSH_ENVIRONMENT
unset BOSH_CA_CERT
unset BOSH_ALL_PROXY
unset CREDHUB_CLIENT
unset CREDHUB_SECRET
unset CREDHUB_CA_CERT
unset CREDHUB_PROXY
```