package logs

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

type RollingFileOutput struct {
	fileName string
	file     *os.File
	queue    chan []byte
	nextTime time.Time
	archive  bool
}

func (this *RollingFileOutput) Write(bs []byte) (int, error) {
	this.queue <- bs
	return len(bs), nil
}

func NewRollingFileOutput(fileNamePattern string, archive bool) (*RollingFileOutput, error) {
	output := &RollingFileOutput{
		fileName: fileNamePattern,
		queue:    make(chan []byte, 1000),
		archive:  archive,
	}
	if err := output.newFile(); err != nil {
		return nil, err
	}
	go output.writeLoop()
	return output, nil
}
func (this *RollingFileOutput) setPreTime() {
	year, month, day := time.Now().Date()
	this.nextTime = time.Date(year, month, day,
		0, 0, 0, 0, time.Now().Location()).
		Add(time.Hour * 24)
}

func (this *RollingFileOutput) newFile() error {
	this.setPreTime()

	// Create dirs if needed
	dir := filepath.Dir(this.fileName)
	if err := os.MkdirAll(dir, os.ModeDir|0755); err != nil {
		return err
	}

	newFile, err := os.OpenFile(this.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	this.file = newFile
	return nil
}

// Archive old file if needed.
func (this *RollingFileOutput) archiveOldFile(fileName string, archive string) {
	if archive, ok := Archivers[archive]; ok {
		err := archive(fileName)
		if err != nil {
			log.Printf("Error on archiving file [%s]: %v\n", fileName, err)
		}
	}
}

func (this *RollingFileOutput) write(bs []byte) error {
	if time.Now().After(this.nextTime) {
		_ = this.file.Close()
		newFileName := this.fileName + "." + this.nextTime.Add(time.Hour*-24).Format("20060102")
		_ = os.Rename(this.fileName, newFileName)
		if this.archive {
			go this.archiveOldFile(newFileName, GzipSuffix)
		}
		_ = this.newFile()
	}

	if _, err := this.file.Write(bs); err != nil {
		return err
	}
	return nil
}

func (this *RollingFileOutput) writeLoop() {
	for entry := range this.queue {
		if err := this.write(entry); err != nil {
			log.Printf("Error on writing to file: %v\n", err)
		}
	}
}
