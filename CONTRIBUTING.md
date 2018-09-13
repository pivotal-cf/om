# Sign the CLA

If you have not previously done so, please fill out and
submit the [Contributor License Agreement](https://cla.pivotal.io).

# Contributing to om

All kinds of contributions to `om` are welcome, whether they be improvements to
documentation, feedback or new ideas, or improving the application directly with
bug fixes, improvements to existing features or adding new features.

## Start with a github issue

In all cases, following this workflow will help all contributors to `om` to
participate more equitably:

1. Search existing github issues that may already describe the idea you have.
   If you find one, consider adding a comment that adds additional context about
   your use case, and/or your interest in helping to contribute to that effort.
2. If there is no existing issue that covers your idea, open a new issue to
   describe the change you would like to see in `om`. Please provide as much
   context as you can about your use case and the reason why you would like to see
   this change. If you are reporting a bug, please include steps to reproduce the
   issue if possible.
3. Any number of folks from the community may comment on your issue and ask
   additional questions. A maintainer will add the `pr welcome` label to the
   issue when it has been determined that the change will be welcome. Anyone
   from the community may step in to make that change.
4. If you intend to make the changes, comment on the issue to indicate your
   interest in working on it to reduce the likelihood that more than one person
   starts to work on it independently.

# Developing Om

## Getting Started

First things first. Just clone the repo and run the tests to make sure you're
ready to safely start exploring or adding new features.

### Clone the repo

```bash
git clone https://github.com/pivotal-cf/om
```

### Run the tests

`om` uses the [ginkgo](https://onsi.github.io/ginkgo/) test framework.

No special `bin` or `scripts` dir here, we run the tests with this one-liner:

```bash
ginkgo -r -race -p .
```

## Vendoring dependencies

The project currently checks in all vendored dependencies. Our vendoring tool of choice
at present is [dep](https://github.com/golang/dep) which is rapidly becoming the standard.

Adding a dependency is relatively straightforward (first make sure you have the dep binary):

```go
  dep ensure -add github.com/some-user/some-dep
```

Check in both the manifest changes and the file additions in the vendor directory.

## Contibuting your changes

1. When you have a set of changes to contribue back to `om`, create a pull
   request (PR) and reference the issue that the changes in the PR are
   addressing.
   **NOTE:** maintainers of `om` with commit access _may_ commit
   directly to `om` instead of creating a pull request. Alternatively, they may choose
   to create a pull request for greater visibility around a set of changes.
   There is no black and white rule here for maintainer. Use your judgement.
2. The code in your pull request will be automatically tested in our continuous
   integration pipeline. At this time, we cannot expose all the logs for this
   pipeline, but we may do so in the future if we can determine it is safe and
   unlikely to lead to any exposure of sensitive information.
3. Your pull request will be reviewed by one or more maintainers. You may also
   receive feedback from others in the community. The feedback may come in the
   form of requests for additional changes to meet expectations for code
   quality, consistency or test coverage. Or it could be clarifying questions to
   better understand the decisions you made in your implementation.
4. When a maintainer accepts your changes, they will merge your pull request.
   If there are outstanding requests for changes or other small changes they
   feel can be made to improve the changed code, they may make additional
   changes or merge the changes manually. It's always nice to have changes come
   in just as the team would like to see them, but we'll try not to hold up a pull
   request for a long period of time due to minor changes.

NOTE: With any significant change in behavior to `om` that should be noted in
the next release's release notes, you should also add a note to [CHANGELOG.md](./CHANGELOG.md).

## Design Goals

- a [small sharp tool](https://brandur.org/small-sharp-tools) for fast, reliable interaction with the Operations Manager API via the command line
- enable humans easily interact with Operations Manager via the command line
- enable scripts and continuous integration systmes to programmatically interact with Operations Manager
- single binary that can be run on multiple platforms without additional dependencies
- a consistent, tested code base that welcomes contributions from the community
- idempotency for all commands

Maintaining _complete_ parity with the features that the Operations Manager API
supports is _not_ an explicit design goal, but we welcome any ideas for new
features that may become feasible as new features are made available in the API.

## Technical Design Guidelines

In general, features are driven by acceptance and unit tests. Acceptance tests execute the compiled binary and exercise the feature including user facing errors. Unit tests document and specify the behavior of smaller components that make up that feature. Take a look at the code around and be consistent. Feel free
to ask questions along the way or to create a pull request early to get feedback
on code that is a work in progress.

# Becoming a committer

At this time, there is no official process for becoming a comitter to `om`.  The
project is currently jointly maintained by Pivotal's PAS Release Engineering
team and Pivotal's Platform Automation team. But we're open to new ideas here!

# Prior Art

`om` was intially developed to be a less flakey / faster replacement of [opsmgr](https://github.com/pivotal-cf/opsmgr)
