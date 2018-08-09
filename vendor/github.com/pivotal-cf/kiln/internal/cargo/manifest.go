package cargo

type Manifest struct {
	Name           string          `yaml:"name"`
	Releases       []Release       `yaml:"releases"`
	Stemcells      []Stemcell      `yaml:"stemcells"`
	Update         Update          `yaml:"update"`
	Variables      []Variable      `yaml:"variables"`
	InstanceGroups []InstanceGroup `yaml:"instance_groups"`
}

type Release struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Stemcell struct {
	Alias   string `yaml:"alias"`
	OS      string `yaml:"os"`
	Version string `yaml:"version"`
}

type Update struct {
	Canaries        int    `yaml:"canaries"`
	CanaryWatchTime string `yaml:"canary_watch_time"`
	UpdateWatchTime string `yaml:"update_watch_time"`
	MaxInFlight     int    `yaml:"max_in_flight"`
	MaxErrors       int    `yaml:"max_errors"`
	Serial          bool   `yaml:"serial"`
}

type Variable struct {
	Name    string      `yaml:"name"`
	Options interface{} `yaml:"options"`
	Type    string      `yaml:"type"`
}

type InstanceGroup struct {
	Name       string             `yaml:"name"`
	AZs        []string           `yaml:"azs"`
	Lifecycle  string             `yaml:"lifecycle"`
	Stemcell   string             `yaml:"stemcell"`
	Instances  int                `yaml:"instances"`
	Jobs       []InstanceGroupJob `yaml:"jobs"`
	Properties interface{}        `yaml:"properties"`
}

type InstanceGroupJob struct {
	Name       string      `yaml:"name"`
	Release    string      `yaml:"release"`
	Provides   interface{} `yaml:"provides"`
	Consumes   interface{} `yaml:"consumes"`
	Properties interface{} `yaml:"properties"`
}
