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
at present is [gvt](https://github.com/FiloSottile/gvt)

Adding a dependency is relatively straightforward (first make sure you have the gvt binary):

```go
  gvt fetch github.com/some-user/some-repo
```

Check in both the manifest changes and the file additions in the vendor directory.

## Running the tests

No special `bin` or `scripts` dir here, we run the tests with this one-liner

```bash
  ginkgo -r -race -p .
```
