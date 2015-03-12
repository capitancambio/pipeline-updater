package updater

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

//Downloader contract
type Downloader interface {
	Download(o io.Writer) error
}

//Struct that contains the information about an artifact
type Artifact struct {
	Id         string //the artifact id
	Href       string //Artifact address
	Version    string //version
	DeployPath string //relative path where to copy the artifact file
}

//downloads the artifact from href
func (a Artifact) Download(w io.Writer) error {
	//check sanity
	if a.DeployPath == "" {
		return fmt.Errorf("DeployPath not set")
	}
	if a.Href == "" {
		return fmt.Errorf("No Href not set")
	}
	resp, err := http.Get(a.Href)
	if err != nil {
		return err
	}
	if resp.StatusCode > 300 {
		return fmt.Errorf("Server %v returned an invalid status %d", a.Href, resp.StatusCode)
	}
	_, err = io.Copy(w, resp.Body)
	return err
}

//An artifact that is present in the local fs
type LocalArtifact struct {
	Artifact
	Path string
}

//Removes this copy of the artifact
func (la LocalArtifact) Clean() error {
	return os.Remove(la.Path)

}

//Copies the artifact having as root directory the path
func (la LocalArtifact) Copy(path string) error {
	absolute := filepath.Join(path, la.DeployPath)
	os.MkdirAll(filepath.Dir(absolute), 0755)
	out, err := os.Create(absolute)
	if err != nil {
		return err
	}
	defer out.Close()
	in, err := os.Open(la.Path)
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = io.Copy(out, in)
	//out.Seek(0, 0)
	//data, err := ioutil.ReadAll(out)
	//fmt.Printf("data %+v\n", string(data))
	return err

}

//convienice struct for storing download results
type downloadResult struct {
	la  LocalArtifact
	err error
}

//downloads the artifacts to the given path
func Download(path string, as ...Artifact) ([]LocalArtifact, error) {
	locals := make([]LocalArtifact, 0, len(as))
	errors := []error{}

	chanArts := make(chan downloadResult)
	//do it async to go faster!!
	for _, artifact := range as {
		//local copy
		a := artifact
		go func() {
			result := downloadResult{
				la: LocalArtifact{
					Artifact: a,
				},
			}

			//create file
			path := filepath.Join(path, a.DeployPath)
			os.MkdirAll(filepath.Dir(path), 0755)
			f, err := os.Create(path)
			if err != nil {
				result.err = err
				chanArts <- result
			}
			log.Println("Downloading ", a.Id, "to", path)
			//download file
			if err := a.Download(f); err != nil {
				result.err = err
				chanArts <- result
			}
			//store the file name
			result.la.Path = f.Name()
			chanArts <- result
		}()

	}
	for i := 0; i < len(as); i++ {
		res := <-chanArts
		if res.err == nil {
			locals = append(locals, res.la)
		} else {
			errors = append(errors, res.err)
		}
	}
	if len(errors) != 0 {
		return []LocalArtifact{}, fmt.Errorf("Errors while downloading %v", errors)
	}
	return locals, nil
}

func Remove(las []LocalArtifact) (ok bool, errs []error) {
	fn := func(l LocalArtifact) error {
		return l.Clean()
	}
	return apply(las, fn)
}
func Copy(las []LocalArtifact, path string) (ok bool, errs []error) {
	fn := func(l LocalArtifact) error {
		return l.Copy(path)
	}
	return apply(las, fn)
}

func apply(las []LocalArtifact, fn func(LocalArtifact) error) (ok bool, errs []error) {
	errs = []error{}
	for _, la := range las {
		if err := fn(la); err != nil {
			errs = append(errs, err)
		}
	}
	return len(errs) == 0, errs
}
