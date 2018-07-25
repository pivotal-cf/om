## 0.39.0 (Unreleased)

BACKWARDS INCOMPATIBILITIES:
- `om interpolate` no longer takes `--output-file` flag.
- `om interpolate` takes ops-files with `-o` instead of `--ops`.
- `om --format=json COMMAND` is no longer supported. This flag should not have
  been global as it is only supported on some commands. The flag is now
  supported on certain commands and needs to be called: `om COMMAND
  --format=json`. The commands that output in `table` format will continue to do
  so by default.

FEATURES:
- `om configure-product` accepts ops-files.
