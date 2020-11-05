// +build tools

/*
	There are binary dependencies of our concourse tasks.
	They need to be declared so their licences can be scanned for OSL.

	If you add something via `go get ...package name...`,
	and it get's remove via `go tidy` add it here.
*/

package binarydeps

import _ "code.cloudfoundry.org/credhub-cli"
import _ "github.com/cloudfoundry/bosh-cli"
import _ "github.com/vmware/govmomi/govc"
import _ "github.com/pivotal-cf/winfs-injector"
import _ "github.com/pivotal-cf/replicator"
