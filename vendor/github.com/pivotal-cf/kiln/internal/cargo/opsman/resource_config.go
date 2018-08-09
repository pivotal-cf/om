package opsman

type ResourceConfig struct {
	Name      string
	Instances ResourceConfigInstances
}

type ResourceConfigInstances struct {
	Value int
}

func (rci ResourceConfigInstances) IsAutomatic() bool {
	return rci.Value < 0
}
