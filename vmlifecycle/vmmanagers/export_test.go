package vmmanagers

import "time"

func (v *VsphereVMManager) SetBootTimeout(timeout time.Duration) {
	v.bootTimeout = timeout
}
