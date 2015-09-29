package core

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/cactus/gologit"
	gcfg "gopkg.in/gcfg.v1"
)

var Config struct{ Main struct{ ZFSRoot string } }

func LoadConfig(filepath string) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		gologit.Fatalf("%s not present or not readable", filepath)
	}

	buffer := &bytes.Buffer{}
	buffer.WriteString("[main]\n")
	f, err := os.Open(filepath)
	if err != nil {
		gologit.Printf("Error reading config file %s", filepath)
		gologit.Fatal(err)
	}
	defer f.Close()

	_, err = buffer.ReadFrom(f)
	if err != nil {
		gologit.Printf("Error reading config file %s", filepath)
		gologit.Fatal(err)
	}

	err = gcfg.ReadInto(&Config, buffer)
	if err != nil {
		gologit.Printf("Error parsing config file %s", filepath)
		gologit.Fatal(err)
	}
}

func GetZFSRootPath() string {
	return Config.Main.ZFSRoot
}

func GetJailsPath() string {
	return filepath.Join(GetZFSRootPath(), "jails")
}
