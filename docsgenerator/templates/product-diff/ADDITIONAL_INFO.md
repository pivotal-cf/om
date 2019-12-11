<!--- Anything in this file will be appended to the final docs/product-diff/README.md file --->
## Important Note
This command is useful for determining the scope of apply changes,
but it has a limitation it's important to be aware of.

The runtime configurations returned by the Ops Manager API
and presented by this command
are those _provided_ by the product.
They are **not** those that _impact_ the product.

For a concrete example, take `p-antivirus`.
It provides runtime configs that put an add-on on every VM.
If such a runtime config has been updated
since the last time the PAS product, `cf`, was deployed,
_every VM in `cf` will be rolled,_
but no runtime config diffs will show if you just run
`om product-diff --product cf`.
To work around this, we recommend you include any add-on products
in your diff command.
For the above example, that would be:
```
om product-diff --product cf --product p-antivirus
```

But even this might not tell you what you need to know!
If the `p-antivirus` changes have already been applied,
they won't show a diff at all.
