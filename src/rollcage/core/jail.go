package core

import (
	"fmt"
	"path"
	"strings"

	"github.com/cactus/gologit"
)

type JailMeta struct {
	Path       string
	Mountpoint string
	HostUUID   string
	Tag        string
	props      ZFSProperties
}

// return whether the jail is running or not
func (jail *JailMeta) IsRunning() bool {
	if jail.GetJID() == "" {
		return false
	}
	return true
}

// return jls jail id for a given jail
// returns emptry string if jail is not running
func (jail *JailMeta) GetJID() string {
	out, err := Jls("-j", fmt.Sprintf("ioc-%s", jail.HostUUID), "jid")
	if err != nil {
		return ""
	}
	return string(out)
}

func (jail *JailMeta) GetProperties() ZFSProperties {
	if jail.props != nil {
		return jail.props
	}
	props := make(ZFSProperties, 0)
	lines := SplitOutput(ZFSMust(
		fmt.Errorf("Error listing properties"),
		"get", "-H", "-o", "property,value", "all", jail.Path))
	if len(lines) < 1 {
		gologit.Fatalf("No output from property fetch\n")
	}
	for _, line := range lines {
		props[strings.TrimSpace(line[0])] = strings.TrimSpace(line[1])
	}
	jail.props = props
	return props
}

func (jail *JailMeta) SetProperties(props ZFSProperties) {
	for key, value := range props {
		ZFSMust(
			fmt.Errorf("Error setting property"),
			"set", fmt.Sprintf("%s=%s", key, value), jail.Path)
		jail.props[key] = value
	}
}

func (jail *JailMeta) GetLogPath() string {
	return path.Join(GetLogDir(), fmt.Sprintf("%s-console.log", jail.HostUUID))
}

func GetAllJails() []*JailMeta {
	list := make([]*JailMeta, 0)
	out := ZFSMust(
		fmt.Errorf("No jails found"),
		"list", "-H",
		"-o", "name,org.freebsd.iocage:host_hostuuid,org.freebsd.iocage:tag,mountpoint",
		"-d", "1", GetJailsPath())
	lines := SplitOutput(out)
	// discard first line, as that is the jail dir itself
	for _, line := range lines[1:] {
		list = append(list, &JailMeta{
			Path:       line[0],
			HostUUID:   line[1],
			Tag:        line[2],
			Mountpoint: line[3],
		})
	}
	return list
}

func FindJail(lookup string) (*JailMeta, error) {
	out, err := ZFS(
		"list", "-H", "-d", "1",
		"-o", "name,org.freebsd.iocage:host_hostuuid,org.freebsd.iocage:tag,mountpoint",
		GetJailsPath())
	if err != nil {
		return nil, err
	}
	lines := SplitOutput(out)
	for _, line := range lines {
		if line[2] == lookup || strings.HasPrefix(line[1], lookup) {
			return &JailMeta{
				Path:       line[0],
				HostUUID:   line[1],
				Tag:        line[2],
				Mountpoint: line[3],
			}, nil
		}
	}
	return nil, fmt.Errorf("No jail found")
}
