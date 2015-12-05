package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cactus/gologit"
)

type ZFSProperties map[string]string

func (prop *ZFSProperties) Get(property string) string {
	return prop[property]
}

func (prop *ZFSProperties) GetIOC(property string) string {
	return prop[strings.Sprintf("org.freebsd.iocage:%s", property)]
}

type JailMeta struct {
	Path     string
	HostUUID string
	Tag      string
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
	props := make(ZFSProperties, 0)
	lines := core.SplitOutput(core.ZFSMust(
		fmt.Errorf("Error listing properties"),
		"get", "-H", "-o", "property,value", jail.Path))
	if len(lines) < 1 {
		gologit.Fatalf("No output from property fetch\n")
	}
	for _, line := range lines {
		props[strings.TrimSpace(line[0])] = strings.TrimSpace(line[1])
	}
	return props
}

func (jail *JailMeta) SetProperties(props ZFSProperties) {
	for key, value := range props {
		core.ZFSMust(
			fmt.Errorf("Error setting property"),
			"set", fmt.Sprintf("%s=%s", key, value), jail.Path)
	}
}

func GetAllJails() []*JailMeta {
	list := make([]*JailMeta, 0)
	out := ZFSMust(
		fmt.Errorf("No jails found"),
		"list", "-H",
		"-o", "name,org.freebsd.iocage:host_hostuuid,org.freebsd.iocage:tag",
		"-d", "1", GetJailsPath())
	lines := SplitOutput(out)
	// discard first line, as that is the jail dir itself
	for _, line := range lines[1:] {
		list = append(list, &JailMeta{
			Path:     line[0],
			HostUUID: line[1],
			Tag:      line[2],
		})
	}
	return list
}

func FindJail(lookup string) (*JailMeta, error) {
	out, err := ZFS(
		"list", "-H", "-d", "1",
		"-o", "name,org.freebsd.iocage:host_hostuuid,org.freebsd.iocage:tag",
		GetJailsPath())
	if err != nil {
		return nil, err
	}
	lines := SplitOutput(out)
	for _, line := range lines {
		if line[2] == lookup || strings.HasPrefix(line[1], lookup) {
			return &JailMeta{
				Path:     line[0],
				HostUUID: line[1],
				Tag:      line[2],
			}, nil
		}
	}
	return nil, fmt.Errorf("No jail found")
}

func ZFS(arg ...string) (string, error) {
	return Cmd("/sbin/zfs", arg...)
}

func ZFSMust(errmsg error, arg ...string) string {
	return CmdMust(errmsg, "/sbin/zfs", arg...)
}

func Jls(arg ...string) (string, error) {
	return Cmd("/usr/sbin/jls", arg...)
}

func JlsMust(errmsg error, arg ...string) string {
	return CmdMust(errmsg, "/usr/sbin/jls", arg...)
}

func Cmd(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	gologit.Debugf("cmd: %s\nstdout: %s\nstderr: %s", cmd.Args, stdout, stderr)
	return strings.TrimSpace(stdout.String()), err
}

func CmdMust(errmsg error, name string, arg ...string) string {
	out, err := Cmd(name, arg...)
	if err != nil {
		gologit.Fatalf("Error: %s\n", errmsg)
	}
	return out
}
