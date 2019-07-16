package common

type Storage struct {
	Name string
	Spec StorageSpec
}

type StorageSpec struct {
	StorageType string
	Hosts       []Host
}

type Host struct {
	NodeName     string
	BlockDevices []string
}

/////////////////////////////////
type Data struct {
	Type         string            `json:"type"`
	ResourceType string            `json:"resourceType"`
	Links        map[string]string `json:"links"`
	Data         []Info            `json:"data"`
}

type Info struct {
	Hosts []HostTmp `json:"hosts"`
}

type HostTmp struct {
	NodeName     string `json:"nodeName"`
	BlockDevices []Dev  `json:"blockDevices"`
}

type Dev struct {
	Name       string `json:"name"`
	Size       string `json:"size"`
	Parted     bool   `json:"parted"`
	Filesystem bool   `json:"filesystem"`
	Mount      bool   `json:"mount"`
}
