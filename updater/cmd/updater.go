package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"bitbucket.org/kardianos/osext"

	"github.com/capitancambio/pipeline-updater/updater"
)

const (
	Latest = "latest"
)

var (
	env = os.Getenv("DP2_HOME")
)

var service = flag.String("service", "http://defaultservice.com", "Url of the update service")
var version = flag.String("version", Latest, "Version to update to")
var installDir = flag.String("install-dir", env, "Pipeline install directory")
var localDescriptor = flag.String("descriptor", "", "Current descriptor")

func main() {
	flag.Parse()
	exePath, err := osext.Executable()
	if err != nil {
		updater.Error(err.Error())
		os.Exit(-1)
	}
	logfile, err := os.Create(filepath.Join(filepath.Dir(exePath), "log.txt"))
	if err != nil {
		updater.Error(err.Error())
		os.Exit(-1)
	}
	log.SetOutput(logfile)

	remote, err := LoadRemote(*service, *version)
	if err != nil {
		updater.Error(err.Error())
		log.Println(err)
		os.Exit(-1)
	}
	local, err := LoadLocal(*localDescriptor)
	if err != nil {
		updater.Error(err.Error())
		log.Println(err)
		os.Exit(-1)
	}
	err = remote.UpdateFrom(local, *installDir)
	if err != nil {
		updater.Error(err.Error())
		log.Println(err)
		os.Exit(-1)
	}

}
func LoadRemote(service, version string) (rd updater.ReleaseDescriptor, err error) {
	rd = updater.NewEmptyReleaseDescriptor()
	resp, err := http.Get(fmt.Sprintf("%s/%s", service, version))
	if err != nil {
		return
	}
	if resp.StatusCode > 300 {
		return rd, fmt.Errorf("Invalid status %v", resp.Status)
	}
	err = xml.NewDecoder(resp.Body).Decode(&rd)
	return

}
func LoadLocal(path string) (rd updater.ReleaseDescriptor, err error) {
	rd = updater.NewEmptyReleaseDescriptor()
	if path == "" {
		return
	}
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	err = xml.NewDecoder(f).Decode(&rd)
	return

}
