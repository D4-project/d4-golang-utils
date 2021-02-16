package inputreader

import (
	"bytes"
	"encoding/base64"
	"github.com/rjeczalik/notify"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// FileWatcherReader is an abstraction of a folder watcher
// and behaves like a reader
type FileWatcherReader struct {
	// Folder to watch
	folderfd *os.File
	// Notify Channel
	eic chan notify.EventInfo
	// TearDown channel
	exit chan string
	// Current buffer
	buf []byte
}

// NewFileWatcherReader creates a new FileWatcherReader
func NewFileWatcherReader(f *os.File) (*FileWatcherReader, error) {
	r := &FileWatcherReader{
		folderfd: f,
		eic: make(chan notify.EventInfo, 1),
	}
	return r, nil
}

// Read  waits for InCloseWrite file event  uses a bytes reader to copy
// the resulting file in p
func (fw *FileWatcherReader) Read(p []byte) (n int, err error) {
	if err := notify.Watch("./...", fw.eic, notify.InCloseWrite); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(fw.eic)

	for{
		select{
			case ei := <-fw.eic:
				// New File, let's read its content
				var err error
				fw.buf, err = ioutil.ReadFile(ei.Path())
				if err != nil {
					log.Fatal(err)
				}
				// base64 stream encoder
				b64buf := new(bytes.Buffer)
				b64encoder := base64.NewEncoder(base64.StdEncoding, b64buf)
				// Encode in Base64 to b64buf
				b64encoder.Write(fw.buf)
				// Close the encoder to flush partially written blocks
				b64encoder.Close()
				b64buf.WriteString("\n")
				//rreader := bytes.NewReader(fw.buf)
				n, err = b64buf.Read(p)
				return n, err
			case <-fw.exit:
				// Exiting
				return 0, io.EOF
		}
	}
}

// Teardown is called on error to stop the Reading loop if needed
func (rl *FileWatcherReader) Teardown() {
	rl.exit <- "exit"
}