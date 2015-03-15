package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

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
	remote, err := LoadRemote(*service, *version)
	if err != nil {
		log.Println(err)
	}
	local, err := LoadLocal(*localDescriptor)
	if err != nil {
		log.Println(err)
	}
	err = remote.UpdateFrom(local, *installDir)
	if err != nil {
		log.Println(err)
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
