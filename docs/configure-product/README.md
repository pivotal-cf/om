&larr; [back to Commands](../README.md)

# `om configure-product`

The `configure-product` command will configure your product properties, network, and resources on the Ops Manager.

## Command Usage
```
ॐ  configure-product
This authenticated command configures a staged product

Usage: om [options] configure-product [<args>]
  -v, --version              bool    prints the om release version (default: false)
  -h, --help                 bool    prints this usage information (default: false)
  -t, --target               string  location of the Ops Manager VM
  -u, --username             string  admin username for the Ops Manager VM (not required for unauthenticated commands)
  -p, --password             string  admin password for the Ops Manager VM (not required for unauthenticated commands)
  -k, --skip-ssl-validation  bool    skip ssl certificate validation during http requests (default: false)
  -r, --request-timeout      int     timeout in seconds for HTTP requests to Ops Manager (default: 1800)

Command Arguments:
  -n, --product-name        string  name of the product being configured
  -p, --product-properties  string  properties to be configured in JSON format (default: )
  -pn, --product-network    string  network properties in JSON format (default: )
  -pr, --product-resources  string  resource configurations in JSON format (default: {})
```

### Configuring the `--product-network`

#### Example JSON:
```json
{
  "singleton_availability_zone": {
    "name": "some-az-1"
  },
  "other_availability_zones": [
    {
      "name": "some-az-1"
    },
    {
      "name": "some-az-2"
    },
    {
      "name": "some-az-3"
    }
  ],
  "network": {
    "name": "some-ert-subnet"
  }
}
```

#### Configuring the `--product-network` on Azure
The product network on Azure does not include Availability Zones, but the API will still expect them to be provided.
To satisfy the API, you can submit "null" AZs for the API as is shown here:
```json
{
  "singleton_availability_zone": {
    "name": "null"
  },
  "other_availability_zones": [
    {
      "name": "null"
    }
  ],
  "network": {
    "name": "example-ert-subnet"
  }
}
```

### Configuring the `--product-properties`
Here is an example of how you might configure the Elastic Runtime Tile.
For the current configuration of your product, you can `curl` the API to retrieve the product properties, eg. `om -u user -p password curl --path /api/v0/staged/products/some-product-guid/properties`.

#### Example JSON:
```json
{
  ".cloud_controller.system_domain": {
    "value": "sys.example.com"
  },
  ".cloud_controller.apps_domain": {
    "value": "apps.example.com"
  },
  ".ha_proxy.skip_cert_verify": {
    "value": true
  },
  ".properties.networking_point_of_entry": {
    "value": "external_ssl"
  },
  ".properties.networking_point_of_entry.external_ssl.ssl_rsa_certificate": {
    "value": {
      "cert_pem": "-----BEGIN CERTIFICATE-----\nSECRET STUFF\n-----END CERTIFICATE-----\n",
      "private_key_pem": "-----BEGIN RSA PRIVATE KEY-----\nSECRET STUFF\n-----END RSA PRIVATE KEY-----\n"
    }
  },
  ".properties.security_acknowledgement": {
    "value": "X"
  },
  ".properties.system_blobstore": {
    "value": "external_gcs"
  },
  ".properties.system_blobstore.external_gcs.packages_bucket": {
    "value": "env-packages"
  },
  ".properties.system_blobstore.external_gcs.droplets_bucket": {
    "value": "env-droplets"
  },
  ".properties.system_blobstore.external_gcs.resources_bucket": {
    "value": "env-resources"
  },
  ".properties.system_blobstore.external_gcs.buildpacks_bucket": {
    "value": "env-buildpacks"
  },
  ".properties.system_blobstore.external_gcs.access_key": {
    "value": "some-access-key"
  },
  ".properties.system_blobstore.external_gcs.secret_key": {
    "value": {
      "secret": "some-secret-key"
    }
  },
  ".properties.tcp_routing": {
    "value": "enable"
  },
  ".properties.tcp_routing.enable.reservable_ports": {
    "value": "1024-1123"
  },
  ".properties.smtp_from": {
    "value": "some-user@example.com"
  },
  ".properties.smtp_address": {
    "value": "smtp.example.com"
  },
  ".properties.smtp_port": {
    "value": "587"
  },
  ".properties.smtp_credentials": {
    "value": {
      "identity": "some-smtp-username",
      "password": "some-smtp-password"
    }
  },
  ".properties.smtp_enable_starttls_auto": {
    "value": true
  },
  ".properties.smtp_auth_mechanism": {
    "value": "plain"
  }
}
```

### Configuring the `--product-resources`

#### Example JSON:
```json
{
  "tcp_router": {
    "elb_names": [
      "some-tcp-load-balancer"
    ]
  },
  "mysql": {
    "instances": 3
  },
  "router": {
    "instances": 3,
    "elb_names": [
      "some-http-load-balancer",
      "some-web-socket-load-balancer"
    ]
  },
  "consul_server": {
    "instances": 3
  },
  "etcd_tls_server": {
    "instances": 3
  },
  "diego_brain": {
    "elb_names": [
      "some-ssh-load-balancer"
    ]
  },
  "diego_cell": {
    "instances": 3
  },
  "diego_database": {
    "instances": 3
  },
  "mysql_proxy": {
    "instances": 2
  }
}
```
