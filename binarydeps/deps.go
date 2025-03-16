//go:build tools
// +build tools

/*
	There are binary dependencies of our concourse tasks.
	They need to be declared so their licences can be scanned for OSL.

	If you add something via `go get ...package name...`,
	and it get's remove via `go tidy` add it here.
*/

package binarydeps

import (
	_ "code.cloudfoundry.org/credhub-cli"
	_ "github.com/cloudfoundry/bosh-cli/v7"
	_ "github.com/pivotal-cf/replicator"
	_ "github.com/pivotal-cf/winfs-injector"
	_ "github.com/vmware/govmomi/govc"
)
