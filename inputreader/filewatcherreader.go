package inputreader

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/rjeczalik/notify"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"os"
	"time"
)

// FileWatcherReader is an abstraction of a folder watcher
// and behaves like a reader
type FileWatcherReader struct {
	// Folder to watch
	folderstr string
	// Notify Channel
	eic chan notify.EventInfo
	// TearDown channel
	exit chan string
	// Chan used to restart the watching channel on a new folder
	dailySwitch     chan bool
	dailySwitchExit chan bool
	// Current buffer
	json bool
	// Current file
	curfile *os.File
	// Current state Watching / Reading
	watching bool
	// Insert Separator
	insertsep bool
	// logging
	log * log.Logger
}

// NewFileWatcherReader creates a new FileWatcherReader
// json specifies whether we now we handle json files
func NewFileWatcherReader(f string, j bool, daily bool, logger *log.Logger) (*FileWatcherReader, error) {
	r := &FileWatcherReader{
		folderstr:       f,
		eic:             make(chan notify.EventInfo, 4096),
		dailySwitch:     make(chan bool),
		dailySwitchExit: make(chan bool),
		json:            j,
		watching:        true,
		insertsep:       false,
		log: logger,
	}
	// go routine holding the watcher
	go setUpWatcher(r, daily)

	// cron task to add daily folder to watch
	if daily {
		c := cron.New()
		c.AddFunc("@midnight", func() {
		//c.AddFunc("@every 10s", func() {
			// Sending exit signal to setUpWatcher
			r.dailySwitch <- true
			// Waiting for exit signal from setUpWatcher
			<-r.dailySwitchExit
			go setUpWatcher(r, daily)
		})
		c.Start()
	}
	return r, nil
}

// setUpWatcher holds the watcher
func setUpWatcher(r *FileWatcherReader, daily bool) {
	if daily {
		// TODO make it customizable
		t, _ := time.ParseDuration("1s")
		retryWatch(r, t)
	} else {
		if err := notify.Watch(fmt.Sprintf("%s/...", r.folderstr), r.eic, notify.InCloseWrite); err != nil {
			log.Fatal(err)
		}
	}
	defer notify.Stop(r.eic)
	<-r.dailySwitch
	r.dailySwitchExit <- true
}

// retryWatch tries to set up the watcher until it works every t
func retryWatch(r *FileWatcherReader, t time.Duration) {
	dt := time.Now()
	//Format YYYYMMDD
	// TODO make it customizable
	currentFolder := dt.Format("20060102")
	r.log.Println(fmt.Sprintf("Watching: %s/%s/...", r.folderstr, currentFolder))
	for {
		if err := notify.Watch(fmt.Sprintf("%s/%s/...", r.folderstr, currentFolder), r.eic, notify.InCloseWrite); err != nil {
			r.log.Println(fmt.Sprintf("Waiting for: %s/%s/... to exist", r.folderstr, currentFolder))
			time.Sleep(t)
		}else{
			return
		}
	}
}

// Read  waits for InCloseWrite file event uses a bytes reader to copy
// the resulting file encoded in b64 in p
func (fw *FileWatcherReader) Read(p []byte) (n int, err error) {

	// Watching for new files to read
	if fw.watching {
	watchloop:
		for {
			select {
			case ei := <-fw.eic:
				//r.log.Println("Got event:", ei)
				// New File, let's read its content
				var err error
				fw.curfile, err = os.Open(ei.Path())
				if err != nil {
					fw.log.Fatal(err)
				}
				fw.watching = false
				break watchloop
			case <-fw.exit:
				// Exiting
				return 0, io.EOF
			}
		}
	}

	// Inserting separator
	if fw.insertsep {
		var buf []byte
		buf = append(buf, "\n"...)
		rreader := bytes.NewReader(buf)
		n, err = rreader.Read(p)
		fw.watching = true
		fw.insertsep = false
		fw.curfile.Close()
		return n, err
	}

	// Reading
	// if not json it could be anything so we encode it in b64
	if !fw.json {
		// base64 stream encoder
		b64buffer := new(bytes.Buffer)
		b64encoder := base64.NewEncoder(base64.StdEncoding, b64buffer)
		// max buffer size is then 3072 = 4096/4*3
		buf := make([]byte, 3072)
		bytesread, err := fw.curfile.Read(buf)
		// buf is the input
		b64encoder.Write(buf[:bytesread])
		// Close the encoder to flush partially written blocks
		b64encoder.Close()
		if err == io.EOF {
			fw.insertsep = true
		}
		// Copy from b64buffer to p
		n, err = b64buffer.Read(p[:len(b64buffer.Bytes())])
	} else {
		n, err = fw.curfile.Read(p)
		if err == io.EOF {
			fw.insertsep = true
		}
	}
	return n, err
}

// Teardown is called on error to stop the Reading loop if needed
func (rl *FileWatcherReader) Teardown() {
	rl.exit <- "exit"
}
