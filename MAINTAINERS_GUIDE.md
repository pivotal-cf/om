# Maintainer's Guide Notes

## Release Process

We use [GoReleaser]() to do the heavy lifing when releasing new versions of [om](https://github.com/pivotal-cf/om).

It will create a new GitHub release, attach the binaries, and update the Homebrew formula.

Here's a quick rundown of the process:

1. Make sure you have the latest version of GoReleaser installed. See [install instructions](https://goreleaser.com/install/).

2. Ensure you have explorted your GitHub token as an environment variable.
    ```
    export GITHUB_TOKEN=<your_token>
    ```
3. Create a new tag for the release. 
    ```
    git tag -a <Major.Minor.Patch> -m "Release v<version>"
    git push origin <Major.Minor.Patch>
    ```

4. Run GoReleaser to create the release.
    ```
    goreleaser release
    ```

5. Edit the newly created release on [GitHub](https://github.com/pivotal-cf/om/releases). There's feature to automatically create release notes between two tags. Use it and then clean up redundant or uninformative entries. See these [instructions](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository#editing-a-release) for more information.
