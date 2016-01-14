package core

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"
)

var jailConfigTemplate = template.Must(template.New("jail.conf").Parse(`
ioc-{{ .UUID }} {
    ip4="{{ .Props.GetIOC "ip4" }}";
    ip4.addr={{range $index, $elem := .IP4 }}{{if $index}}, {{end}}"{{$elem}}"{{ else }}""{{ end }};
    ip4.saddrsel="{{ .Props.GetIOC "ip4_saddrsel" }}";
    ip6="{{ .Props.GetIOC "ip6" }}";
    ip6.addr={{range $index, $elem := .IP6 }}{{if $index}}, {{end}}"{{$elem}}"{{ else }}""{{ end }};
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

func (jail *JailMeta) JailConfig() string {
	props := jail.GetProperties()

	data := struct {
		Props   ZFSProperties
		LogPath string
		Fstab   string
		Root    string
		UUID    string
		IP4     []string
		IP6     []string
	}{
		Props:   props,
		LogPath: jail.GetLogPath(),
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
			ip4 := props.GetIOC("ip4_addr")
			if ip4 != "" {
				ns := []string{}
				s := strings.Split(ip4, ",")
				for _, x := range s {
					ns = append(ns, strings.TrimSpace(x))
				}
				data.IP4 = ns
			}
		}
		if props.GetIOC("ip6_addr") != "none" {
			ip6 := props.GetIOC("ip6_addr")
			if ip6 != "" {
				ns := []string{}
				s := strings.Split(ip6, ",")
				for _, x := range s {
					ns = append(ns, strings.TrimSpace(x))
				}
				data.IP6 = ns
			}
		}
	}

	b := &bytes.Buffer{}
	jailConfigTemplate.Execute(b, data)
	return b.String()
}
