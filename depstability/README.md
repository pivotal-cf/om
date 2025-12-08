# Depstablity Test
This is a test suite intended
to allow any gomod-based go project
to keep track of the impact of dependency changes
on product's ability to reuse OSL files
in the VMware release process.

As a short restatement of VMware policy on this matter,
OSL files may be reused if product's open source dependencies -
_irrespective of version_ -
have not changed.

This means if dependencies were neither added nor removed,
the entire expensive and time consuming OSM process can be bypassed.

As this test is intended to be used to help people
_avoid_ unintended additional dependencies,
it constraints itself to ginkgo, gomega,
and packages from the go standard library.

If you need to update the fixture file,
you can either manually edit it based on the test output,
or you can generate it from the go.sum:

```
cat ../go.sum | awk '{print $1}' | uniq | sort > records/depnames-newversion.txt
```
(The above command should be either run
from this directory,
or modified appropriately for where you run it from.)

If you generate a new one, you'll also need to update
the fixture file in the test.

Since each fixture represents a version,
a new file should be generated for each _released/OSM-requested_ version,
but files for _unreleased_ versions can be updated,
so long as the OSL file hasn't been requested in OSM yet.
