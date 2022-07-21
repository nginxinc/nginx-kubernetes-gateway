package state

import (
	"io/fs"
	"io/ioutil"
	"os"
)

type stdLibFileManager struct{}

func newStdLibFileManager() *stdLibFileManager {
	return &stdLibFileManager{}
}

func (s *stdLibFileManager) ReadDir(dirname string) ([]fs.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

func (s *stdLibFileManager) Remove(name string) error {
	return os.Remove(name)
}

func (s *stdLibFileManager) Write(file *os.File, contents []byte) error {
	_, err := file.Write(contents)

	return err
}

func (s *stdLibFileManager) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func (s *stdLibFileManager) Chmod(file *os.File, mode os.FileMode) error {
	return file.Chmod(mode)
}
