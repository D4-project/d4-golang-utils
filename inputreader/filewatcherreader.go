package inputreader

import (
	"bytes"
	"github.com/gomodule/redigo/redis"
	"github.com/rjeczalik/notify"
	"io"
	"log"
	"os"
)

// FileWatcherReader is a abstraction a folder watcher
// and behaves like a reader
type FileWatcherReader struct {
	// Folder to watch
	folderfd os.File
	// Notify Channel
	eic chan notify.EventInfo
	// Current buffer
	buf []byte
}

// NewLPOPReader creates a new RedisLPOPReader
func NewFileWatcherReader(f os.File) (*FileWatcherReader, error) {
	r := &FileWatcherReader{
		folderfd: f,
		eic: make(chan notify.EventInfo, 1),
	}

	return r, nil
}

// Read  waits for new file event  uses a bytes reader to copy
// the resulting file in p
func (fw *FileWatcherReader) Read(p []byte) (n int, err error) {
	if err := notify.Watch("./...", fw.eic, notify.Remove); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(fw.eic)

	// Create a go routing listening the the channel

	// select, on new event, stream content of the file

		// Block until event is received
		ei := <-fw.eic
		log.Println("Got event:", ei)

	// TODO grab the ei.Path content

	// Encode it in base64
	// push in buffer
	// add \n

	//buf = append(buf, "\n"...)
	//rreader := bytes.NewReader(buf)
	//n, err = rreader.Read(p)
	//return n, err
}

// Teardown is called on error to close the redis connection
func (rl *RedisLPOPReader) Teardown() {
	(*rl.r).Close()
}