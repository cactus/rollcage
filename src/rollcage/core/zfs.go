package core

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/cactus/gologit"
)

type JailMeta struct {
	Path     string
	HostUUID string
	Tag      string
}

func GetJailId(hostUUID string) ([]byte, error) {
	cmd := exec.Command("/usr/sbin/jls", "-j",
		fmt.Sprintf("ioc-%s", hostUUID), "jid")
	gologit.Debugln(cmd.Args)
	out, err := cmd.CombinedOutput()
	return out, err
}

func Jls(arg ...string) ([]byte, error) {
	cmd := exec.Command("/usr/sbin/jls", arg...)
	gologit.Debugln(cmd.Args)
	out, err := cmd.CombinedOutput()
	return out, err
}

func JlsMust(arg ...string) []byte {
	return CmdErrExit(Jls(arg...))
}

func GetAllJails() []*JailMeta {
	list := make([]*JailMeta, 0)
	out := ZFSMust("list", "-H", "-o", "name,org.freebsd.iocage:host_hostuuid,org.freebsd.iocage:tag", "-d", "1", GetJailsPath())
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

func ZFS(arg ...string) ([]byte, error) {
	cmd := exec.Command("/sbin/zfs", arg...)
	gologit.Debugln(cmd.Args)
	out, err := cmd.CombinedOutput()
	return out, err
}

func ZFSMust(arg ...string) []byte {
	return CmdErrExit(ZFS(arg...))
}

func CmdMust(out []byte, err error) []byte {
	return CmdErrExit(out, err)
}

func CmdErrExit(out []byte, err error) []byte {
	if err != nil {
		if out != nil {
			gologit.Println(strings.TrimSpace(string(out)))
		}
		gologit.Fatal(err)
	}
	return out
}
