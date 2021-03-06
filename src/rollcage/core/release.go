package core

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

var SupportedReleases = []string{
	"10.2-RELEASE",
	"10.1-RELEASE",
	"9.3-RELEASE",
}

type ReleaseMeta struct {
	Path       string
	Mountpoint string
	Name       string
	Patchlevel string
}

func GetAllReleases() []*ReleaseMeta {
	list := make([]*ReleaseMeta, 0)
	out := ZFSMust(
		fmt.Errorf("No releases found"),
		"list", "-H", "-d", "1",
		"-o", "name,mountpoint",
		GetReleasesPath())
	lines := SplitOutput(out)
	// discard first line, as that is the jail dir itself
	for _, line := range lines[1:] {
		patchlevel := "??"
		fpath := path.Join(line[1], "root/bin/freebsd-version")
		if _, err := os.Stat(fpath); !os.IsNotExist(err) {
			out, err := exec.Command(fpath).Output()
			if err == nil && len(out) > 0 {
				patchlevel = strings.TrimSpace(string(out))
			}
		}
		list = append(list, &ReleaseMeta{
			Path:       line[0],
			Mountpoint: line[1],
			Name:       path.Base(line[1]),
			Patchlevel: patchlevel,
		})
	}
	return list
}

func FindRelease(lookup string) (*ReleaseMeta, error) {
	out, err := ZFS(
		"list", "-H", "-d", "1",
		"-o", "name,mountpoint",
		GetReleasesPath())
	if err != nil {
		return nil, err
	}

	lines := SplitOutput(out)
	for _, line := range lines {
		name := path.Base(line[1])
		if name == lookup || strings.HasPrefix(name, lookup) {
			patchlevel := "??"
			fpath := path.Join(line[1], "root/bin/freebsd-version")
			if _, err := os.Stat(fpath); !os.IsNotExist(err) {
				out, err := exec.Command(fpath).Output()
				if err == nil && len(out) > 0 {
					patchlevel = strings.TrimSpace(string(out))
				}
			}
			return &ReleaseMeta{
				Path:       line[0],
				Mountpoint: line[1],
				Name:       name,
				Patchlevel: patchlevel,
			}, nil
		}
	}
	return nil, fmt.Errorf("No release found")
}

func CreateRelease(relname string) (*ReleaseMeta, error) {
	releaseName := strings.ToUpper(relname)

	release, _ := FindRelease(relname)
	if release != nil {
		return release, nil
	}

	// create zfs
	basepath := path.Join(GetReleasesPath(), releaseName)
	rootpath := path.Join(basepath, "root")
	setspath := path.Join(basepath, "sets")

	_, err := ZFS("create", "-o", "compression=lz4", "-p", rootpath)
	if err != nil {
		return nil, fmt.Errorf("Error creating release zfs filesystem")
	}

	_, err = ZFS("create", "-o", "compression=off", "-p", setspath)
	if err != nil {
		return nil, fmt.Errorf("Error creating release sets zfs filesystem")
	}

	baseMountPoint := ZFSMust(
		fmt.Errorf("Error getting property in create release"),
		"get", "-H", "-o", "value", "mountpoint", basepath)

	release = &ReleaseMeta{
		Name:       releaseName,
		Path:       basepath,
		Mountpoint: baseMountPoint,
	}

	return release, nil
}
