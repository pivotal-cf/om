### Configuring via YAML config file

The preferred approach is to include all configuration in a single YAML
configuration file.

#### Example YAML

You can turn on or off errands for products you are deploying (or set it to `default`):

```yaml
errands:
  sample-product:
    run_post_deploy:
      smoke_tests: default
      push-usage-service: false
    run_pre_delete:
      smoke_tests: true
  test-product:
    run_post_deploy:
      smoke_tests: default
```

To retrieve the default configuration of your product's errands you can use the `om
staged-config` command (although the returned shape is different).