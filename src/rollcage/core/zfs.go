package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cactus/gologit"
)

type ZFSProperties map[string]string

func (prop ZFSProperties) Get(property string) string {
	return prop[property]
}

func (prop ZFSProperties) GetIOC(property string) string {
	return prop[fmt.Sprintf("org.freebsd.iocage:%s", property)]
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
