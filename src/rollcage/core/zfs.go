package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/cactus/gologit"
)

var JailTemplate = template.Must(template.New("jail.conf").Parse(`
ioc-{{ .UUID }} {
    ip4="{{ .Props.GetIOC "ip4" }}";
    ip4.addr="{{ .IP4 }}";
    ip4.saddrsel="{{ .Props.GetIOC "ip4_saddrsel" }}";
    ip6="{{ .Props.GetIOC "ip6" }}";
    ip6.addr="{{ .IP6 }}";
    ip6.saddrsel="{{ .Props.GetIOC "ip6_saddrsel" }}";
    host.hostname="{{ .Props.GetIOC "hostname" }}";
    host.hostuuid="{{ .Props.GetIOC "host_hostuuid" }}";
    path="{{ .Root }}";
    securelevel="{{ .Props.GetIOC "securelevel" }}";
    devfs_ruleset="{{ .Props.GetIOC "devfs_ruleset" }}";
    enforce_statfs="{{ .Props.GetIOC "enforce_statfs" }}";
    children.max="{{ .Props.GetIOC "children_max" }}";
    allow.set_hostname="{{ .Props.GetIOC "allow_set_hostname" }}";
    allow.sysvipc="{{ .Props.GetIOC "allow_sysvipc" }}";
    allow.chflags="{{ .Props.GetIOC "allow_chflags" }}";
    allow.mount="{{ .Props.GetIOC "allow_mount" }}";
    allow.mount.devfs="{{ .Props.GetIOC "allow_mount_devfs" }}";
    allow.mount.nullfs="{{ .Props.GetIOC "allow_mount_nullfs" }}";
    allow.mount.procfs="{{ .Props.GetIOC "allow_mount_procfs" }}";
    allow.mount.tmpfs="{{ .Props.GetIOC "allow_mount_tmpfs" }}";
    allow.mount.zfs="{{ .Props.GetIOC "allow_mount_zfs" }}";
    mount.fdescfs="{{ .Props.GetIOC "mount_fdescfs" }}";
    allow.quotas="{{ .Props.GetIOC "allow_quotas" }}";
    allow.socket_af="{{ .Props.GetIOC "allow_socket_af" }}";
    mount.fstab="{{ .Fstab }}";
    mount.devfs="{{ .Props.GetIOC "mount_devfs" }}";

    exec.prestart="{{ .Props.GetIOC "exec_prestart" }}";
    exec.start="{{ .Props.GetIOC "exec_start" }}";
    exec.poststart="{{ .Props.GetIOC "exec_poststart" }}";
    exec.prestop="{{ .Props.GetIOC "exec_prestop" }}";
    exec.stop="{{ .Props.GetIOC "exec_stop" }}";
    exec.poststop="{{ .Props.GetIOC "exec_poststop" }}";
    exec.clean="{{ .Props.GetIOC "exec_clean" }}";
    exec.timeout="{{ .Props.GetIOC "exec_timeout" }}";
    exec.fib="{{ .Props.GetIOC "exec_fib" }}";
    stop.timeout="{{ .Props.GetIOC "stop_timeout" }}";
    exec.consolelog="{{ .LogPath }}";
    
    allow.dying;
    persist;
}
`))

type ZFSProperties map[string]string

func (prop ZFSProperties) Get(property string) string {
	return prop[property]
}

func (prop ZFSProperties) GetIOC(property string) string {
	return prop[fmt.Sprintf("org.freebsd.iocage:%s", property)]
}

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

func (jail *JailMeta) JailConfig() string {
	props := jail.GetProperties()

	data := struct {
		Props   ZFSProperties
		LogPath string
		Fstab   string
		Root    string
		UUID    string
		IP4     string
		IP6     string
	}{
		Props:   props,
		LogPath: path.Join(GetLogDir(), fmt.Sprintf("%s-console.log", jail.HostUUID)),
		UUID:    jail.HostUUID,
		Root:    path.Join(jail.Mountpoint, "root"),
		Fstab:   path.Join(jail.Mountpoint, "fstab"),
	}

	if props.GetIOC("vnet") == "on" {
		fmt.Printf("  + Configuring VNET\n")
		// start VNET networking
	} else {
		// start standard networking (legacy?)
		if props.GetIOC("ip4_addr") != "none" {
			data.IP4 = props.GetIOC("ip4_addr")
		}
		if props.GetIOC("ip6_addr") != "none" {
			data.IP6 = props.GetIOC("ip6_addr")
		}
	}

	b := &bytes.Buffer{}
	JailTemplate.Execute(b, data)
	return b.String()
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
