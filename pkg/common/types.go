package common

/*
type Storage struct {
	Name   string
	Spec   StorageSpec
	Status string
}

type StorageSpec struct {
	StorageType string
	Hosts       []Host
}

type Host struct {
	NodeName     string
	BlockDevices []Dev
}

type Dev struct {
	Name string
	Size string
}
*/
/////////////////////////////////
type Data struct {
	Type         string            `json:"type"`
	ResourceType string            `json:"resourceType"`
	Links        map[string]string `json:"links"`
	Data         []HostTmp         `json:"data"`
}

type HostTmp struct {
	NodeName     string   `json:"nodeName"`
	BlockDevices []DevTmp `json:"blockDevices"`
}

type DevTmp struct {
	Name       string `json:"name"`
	Size       string `json:"size"`
	Parted     bool   `json:"parted"`
	Filesystem bool   `json:"filesystem"`
	Mount      bool   `json:"mount"`
}
