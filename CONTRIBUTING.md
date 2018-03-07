# Sign the CLA

If you have not previously done so, please fill out and
submit the [Contributor License Agreement](https://cla.pivotal.io/sign/pivotal).

# Developing Om

## Design Goals

- less flakey / faster replacement of [opsmgr](https://github.com/pivotal-cf/opsmgr)
- single binary that can be run on multiple platforms
- split environment creation from Ops Manager configuration (these are two tools)
- no longer rely on specific environment file format
- fully tested, not using tests to execute browser behavior
- no capybara
- [small sharp tool](https://brandur.org/small-sharp-tools)
- idempotency for all commands

## Vendoring

The project currently checks in all vendored dependencies. Our vendoring tool of choice
at present is [dep](https://github.com/golang/dep) which is rapidly becoming the standard.

Adding a dependency is relatively straightforward (first make sure you have the dep binary):

```go
  dep ensure -add github.com/some-user/some-dep
```

Check in both the manifest changes and the file additions in the vendor directory.

## Running the tests

No special `bin` or `scripts` dir here, we run the tests with this one-liner

```bash
  ginkgo -r -race -p .
```
