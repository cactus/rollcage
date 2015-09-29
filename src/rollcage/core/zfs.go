package core

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/cactus/gologit"
)

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

func GetAllJails() []string {
	out := ZFSMust("list", "-H", "-o", "name", "-d", "1", GetJailsPath())
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	// discard first line, as that is the jail dir itself
	if len(lines) > 0 {
		lines = lines[1:]
	}
	return lines
}

func GetJailUUIDByTagOrUUID(tag string) string {
	out := ZFSMust(
		"list", "-H", "-d", "1",
		"-o", "name,org.freebsd.iocage:host_hostuuid,org.freebsd.iocage:tag",
		GetJailsPath())
	lines := SplitOutput(out)
	for _, line := range lines {
		if line[2] == tag || strings.HasPrefix(line[1], tag) {
			return line[1]
		}
	}
	return ""
}

func GetJailByTagOrUUID(tag string) string {
	out := ZFSMust(
		"list", "-H", "-d", "1",
		"-o", "name,org.freebsd.iocage:host_hostuuid,org.freebsd.iocage:tag",
		GetJailsPath())
	lines := SplitOutput(out)
	for _, line := range lines {
		if line[2] == tag || strings.HasPrefix(line[1], tag) {
			return line[0]
		}
	}
	return ""
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
