## Shell Completion for `om`

`om-completion.sh` provides command completion for the `om` utility and works in
both `bash` and `zsh`. Like most command completion, type `om` and either press tab
twice or begin typing a command and press tab and it will give you the available
options.

### Examples

```sh
$ om <tab><tab>
activate-certificate-authority  create-certificate-authority    errands                         stage-product
apply-changes                   create-vm-extension             export-installation             staged-config
assign-stemcell                 credential-references           generate-certificate            staged-director-config
available-products              credentials                     generate-certificate-authority  staged-manifest
bosh-env                        curl                            help                            staged-products
certificate-authorities         delete-certificate-authority    import-installation             tile-metadata
certificate-authority           delete-installation             installation-log                unstage-product
config-template                 delete-product                  installations                   update-ssl-certificate
configure-authentication        delete-ssl-certificate          interpolate                     upload-product
configure-director              delete-unused-products          pending-changes                 upload-stemcell
configure-ldap-authentication   deployed-manifest               regenerate-certificates         version
configure-product               deployed-products               revert-staged-changes
configure-saml-authentication   download-product                ssl-certificate
```

```sh
$ om c<tab><tab>
certificate-authorities        configure-director             create-certificate-authority   curl
certificate-authority          configure-ldap-authentication  create-vm-extension
config-template                configure-product              credential-references
configure-authentication       configure-saml-authentication  credentials
```
### Usage

#### `bash`
If you already use bash completion, drop this file in `/etc/bash_completion.d` and
it will likely be taken care of. If not, add this file to your `~/.bash_profile`
script:

```
source /path/to/om-completion.sh
```

#### `zsh`
Similarly to `bash`, if you use `zsh`, place the autocomplete script where you
usually keep them, and it should mostly be taken care of. If you don't, put it in
a useful place to you, and add the `source` command to your `.zshrc` file. In either
case, you *must* make sure that the following two lines are executed *before* 
`om-completion.sh` is sourced:

```sh
autoload -U +X compinit && compinit
autoload -U +X bashcompinit && bashcompinit
```
