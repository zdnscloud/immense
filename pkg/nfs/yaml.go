package nfs

import (
	"fmt"

	"github.com/zdnscloud/immense/pkg/common"
)

func csiyaml(name, host, path string) (string, error) {
	cfg := map[string]interface{}{
		"RBACConfig":          common.RBACConfig,
		"StorageNamespace":    common.StorageNamespace,
		"NFSProvisionerImage": NFSProvisionerImage,
		"ProvisionerName":     fmt.Sprintf("%s.%s", name, NfsDriverSuffix),
		"StorageClassName":    name,
		"NfsCSIDpName":        fmt.Sprintf("%s-%s", name, NfsCSIDpSuffix),
		"NfsServer":           host,
		"NfsPath":             path,
	}
	return common.CompileTemplateFromMap(NfsCSITemplate, cfg)
}
func scyaml(name string) (string, error) {
	cfg := map[string]interface{}{
		"StorageClassName": name,
		"ProvisionerName":  fmt.Sprintf("%s.%s", name, NfsDriverSuffix),
	}
	return common.CompileTemplateFromMap(StorageClassTemp, cfg)
}
