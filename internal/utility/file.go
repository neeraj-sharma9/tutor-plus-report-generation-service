package utility

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
)

// ProjectRootPath gets the current_project root-path
func ProjectRootPath() string {
	_, b, _, _ := runtime.Caller(0)
	// remove the current-folder from the current-directory of file
	rootPath := strings.Split(filepath.Dir(b), "")
	return rootPath[0]
}

// GetFileContent reads content from the file and returns the bytes
func GetFileContent(path string) (content []byte, err error) {
	content, err = ioutil.ReadFile(ProjectRootPath() + path)
	if err != nil {
		err = fmt.Errorf("failed to fetch content from file %s : %s", ProjectRootPath()+path, err)
		return
	}
	return
}
