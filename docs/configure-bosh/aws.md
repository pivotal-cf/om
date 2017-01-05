&larr; [back to `configure-bosh`](README.md)

# AWS configure-bosh

#### --iaas-configuration
**Note for `ssh_private_key`**: You will need to replace newlines with "\n".
This can be done with the simple `bash` chainsaw `cat key.pem | awk '{print $1 "\\n"}' | tr -d '\n'`.

##### Minimal example
```json
{
  "access_key_id": "my-access-key",
  "secret_access_key": "my-secret-key",
  "vpc_id": "vpc-123456",
  "security_group": "sg-123456",
  "key_pair_name": "some-key-pair-name",
  "ssh_private_key": "my-private-key",
  "region": "us-west-1",
  "encrypted": false
}

```

#### --director-configuration
**Note regarding `ntp_servers_string`**: We recommend using this NTP server to all PCF users on AWS

##### Minimal example
```json
{
  "ntp_servers_string": "0.amazon.pool.ntp.org"
}
```

#### --security-configuration
No additional security configuration is strictly required.

##### Minimal example
```json
{
  "trusted_certificates": "some-trusted-certificates",
  "vm_password_type": "generate"
}
```

#### --az-configuration

##### Minimal example
```json
{
  "availability_zones": ["us-west-1b","us-west-1c"]
}
```

#### --networks-configuration

##### Minimal example
```json
{
}
```

#### --network-assignment

##### Minimal example
```json
{
}
```
