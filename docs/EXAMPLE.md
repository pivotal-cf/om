&larr; [back to Commands](README.md)

The following example shows how one could configure and deploy the ERT on vSphere.
Secrets and identifying details have been removed. Make sure to include your own
details when following this example.

## 1. Boot your Ops Manager VM.

  This can be done with the existing `opsmgr` tooling.
  For the purposes of this example we will have an Ops Manager VM located at `https://opsman.example.com`.

## 2. Configure the authentication mechanism for the Ops Manager.

This will setup the admin user account. More documentation for the `configure-authentication` command
can be found [here](https://github.com/pivotal-cf/om/blob/master/docs/configure-authentication/README.md).
The command will run, waiting for the UAA to start and creating your admin user account.

**NOTE:** when targetting an Ops Manager that has a self-signed certificate, you should use the
`--skip-ssl-validation` flag to skip validation of the self-signed certificate.

```shell
om \
  --target https://opsman.example.com \
    configure-authentication \
      --user my-user \
      --password my-password \
      --decryption-passphrase my-passphrase
```

## 3. Configure the Director.

This command will fill out the configuration details for the director tile that came with your Ops Manager VM.
The specific configuration for the director changes based on what IAAS you are targetting.
More documentation for the `configure-director` command along with IAAS-specific details can be found
[here](configure-director/README.md).

```shell
om \
  --target https://opsman.example.com \
  --user my-user \
  --password my-password \
    configure-director \
      --config director.yml
```

an example director.yml:
```yaml
iaas-configuration:
  project: my-foo-project
  default_deployment_tag: foo-vms
  auth_json: '{"some-key":"some-value"}'
director-configuration:
  ntp_servers_string: 169.254.169.254
security-configuration:
  trusted_certificates: some-trusted-certificates
az-configuration:
- name: us-central1-a
- name: us-central1-b
- name: us-central1-c
networks-configuration:
  icmp_checks_enabled: false
  networks:
  - name: opsman-network
    subnets:
    - iaas_identifier: some-network/opsman-subnet/us-central1
      cidr: 10.0.0.0/24
      reserved_ip_ranges: 10.0.0.0-10.0.0.4
      dns: 8.8.8.8
      gateway: 10.0.0.1
      availability_zone_names:
      - us-central1-a
      - us-central1-b
      - us-central1-c
network-assignment:
  singleton_availability_zone:
    name: us-central1-a
  network:
    name: opsman-network
```

## 4. Uploading a product.

This command will upload a product file from a local filesystem onto the Ops Manager VM.
More documentation for the `upload-product` command can be found [here](https://github.com/pivotal-cf/om/blob/master/docs/upload-product/README.md).
The command expects a fully qualified path to the product file on the local filesystem.

```shell
om \
  --target https://opsman.example.com \
  --user my-user \
  --password my-password \
    upload-product \
      --product /absolute/path/to/the/product.tgz
```

## 5. Upload a stemcell.

This command will upload a stemcell file from a local filesystem onto the Ops Manager VM.
More documentation for the `upload-stemcell` command can be found [here](https://github.com/pivotal-cf/om/blob/master/docs/upload-stemcell/README.md).
The command expects a fully qualified path to the stemcell file on the local filesystem.

```shell
om \
  --target https://opsman.example.com \
  --user my-user \
  --password my-password \
    upload-stemcell \
      --stemcell /absolute/path/to/the/stemcell.tgz
```

## 6. List available products.

This command is helpful when you are uploading a new product to your Ops Manager.
It will list out the names and versions of all available products.
More documentation for the `available-products` command can be found [here](https://github.com/pivotal-cf/om/blob/master/docs/available-products/README.md).

```shell
om \
  --target https://opsman.example.com \
  --user my-user \
  --password my-password \
    available-products
```

## 7. Stage a product.

This command will move the product from the "Available Products" section of Ops Manager onto the Installation Dashboard.
More documentation for the `stage-product` command can be found [here](https://github.com/pivotal-cf/om/blob/master/docs/stage-product/README.md).

```shell
om \
  --target https://opsman.example.com \
  --user my-user \
  --password my-password \
    stage-product \
      --product-name some-product \
      --product-version 1.2.3-build.4
```

## 8. Configure a product.

This command will specify all of the configuration properties for a product, including network and resource configurations.
More documentation for the `configure-product` command can be found [here](https://github.com/pivotal-cf/om/blob/master/docs/configure-product/README.md).

**NOTE:** If you are looking for the current configuration of your product properties, you can use the `curl` command to
fetch those details. For example, you can use the following command to get the GUIDs for your staged products.

```shell
om \
  --target https://opsman.example.com \
  --user my-user \
  --password my-password \
    curl \
      --path /api/v0/staged/products
```

This will return a JSON list of product and their GUIDs. Once you have a GUID, you can fetch the product properties as follows:

```shell
om \
  --target https://opsman.example.com \
  --user my-user \
  --password my-password \
    curl \
      --path /api/v0/staged/products/some-GUID/properties
```

With this property information in hand, you can choose how to configure the product like this:

```shell
om \
  --target https://opsman.example.com \
  --user my-user \
  --password my-password \
    configure-product \
      --config product.yml
```

an example product.yml:
```yaml
product-name: some-product
network-properties:
  singleton_availability_zone:
    name: az-1
  other_availability_zones:
  - name: az-1
  - name: az-2
  - name: az-3
  network:
    name: ert-network
product-properties:
  ".cloud_controller.system_domain":
    value: sys.example.com
  ".cloud_controller.apps_domain":
    value: apps.example.com
  ".ha_proxy.skip_cert_verify":
    value: true
  ".properties.networking_point_of_entry":
    value: external_ssl
  ".properties.networking_point_of_entry.external_ssl.ssl_rsa_certificate":
    value:
      cert_pem: |
        -----BEGIN CERTIFICATE-----
        SECRET STUFF
        -----END CERTIFICATE-----
      private_key_pem: |
        -----BEGIN RSA PRIVATE KEY-----
        SECRET STUFF
        -----END RSA PRIVATE KEY-----
resource-config:
  tcp_router:
    elb_names:
    - some-tcp-load-balancer
  mysql:
    instances: 3
``` 

## 9. Apply changes.

This command is equivalent to clicking the "Apply Changes" button in the Ops Manager UI. It will deploy your products.
More documentation for the `apply-changes` command can be found [here](https://github.com/pivotal-cf/om/blob/master/docs/apply-changes/README.md).
If the command exits for some sort of networking error, don't worry, your deployment is still running. To connect to
a running installation, just run the command as normal and it will reattach to that deployment.

```shell
om \
  --target https://opsman.example.com \
  --user my-user \
  --password my-password \
    apply-changes
```
