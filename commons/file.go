package commons

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type File struct {
	path string
}

func NewFile(path string) *File {
	absPath, _ := filepath.Abs(path)
	return &File{path: absPath}
}

func (self *File) IsFile() bool {
	return !self.IsDir()
}
func (self *File) IsDir() bool {
	if !self.Exist() {
		return false
	}
	fs, _ := os.Stat(self.path)
	return fs.IsDir()
}
func (self *File) Exist() bool {
	_, err := os.Stat(self.path)
	return err == nil || os.IsExist(err)
}
func (self *File) Name() string {
	return filepath.Base(self.path)
}

func (self *File) Mkdir() error {
	if self.Exist() {
		return nil
	}
	return os.Mkdir(self.path, os.ModePerm)
}

func (self *File) Parent() *File {
	if filepath.Dir(self.path) == self.path {
		return nil
	}
	return NewFile(filepath.Dir(self.path))
}

func (self *File) Equal(file *File) bool {
	return self.path == file.path
}

func (self *File) List() ([]*File, error) {
	dir, err := ioutil.ReadDir(self.path)
	if err != nil {
		return nil, err
	}
	files := make([]*File, len(dir))
	for idx, finfo := range dir {
		files[idx] = NewFile(self.path + "/" + finfo.Name())
	}
	return files, nil
}

func (self *File) Rename(newName string) error {
	dir, _ := filepath.Split(self.path)
	newPath := dir + "/" + newName
	return os.Rename(self.path, newPath)
}

func (self *File) GetPath() string {
	return self.path
}

func (self *File) ToString() (string, error) {
	bs, err := self.ToBytes()
	return string(bs), err
}

//delete file or folder
func (self *File) Remove() error {
	return os.Remove(self.path)
}

//delete file or folder and subfolder
func (self *File) RemoveAll() error {
	return os.RemoveAll(self.path)
}

func (self *File) ToBytes() ([]byte, error) {
	return ioutil.ReadFile(self.path)
}

func (self *File) GetWriter(append bool) (*os.File, error) {
	flag := os.O_RDWR | os.O_CREATE
	if append && self.Exist() {
		flag = flag | os.O_APPEND
	}
	return os.OpenFile(self.path, flag, 0666)
}

func (self *File) Size() int64 {
	if self.IsFile() {
		f, _ := os.Stat(self.path)
		return f.Size()
	} else {
		return -1
	}
}

func (self *File) GetReader() (*os.File, error) {
	if self.Exist() && self.IsFile() {
		return os.OpenFile(self.path, (os.O_RDWR | os.O_APPEND), 0666)
	}
	return nil, errors.New("not found or is not file")
}

func (self *File) Lines() ([]string, error) {
	if r, err := self.GetReader(); err != nil {
		return nil, err
	} else {
		reader := bufio.NewReader(r)
		lines := make([]string, 0)
		for {
			if line, _, err := reader.ReadLine(); err == nil {
				lines = append(lines, string(line))
			} else if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		return lines, nil
	}
}
