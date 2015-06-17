package storage

import (
	"sync"
	"time"
	"runtime"
	"os"
	"encoding/gob"
	"io"
	"fmt"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("example")

type Item struct {
	Value     string
	createdAt int32
}

type Storage struct {
	*storage
}

type storage struct {
	sync.RWMutex
	items  map[string]string
	loader *loader
}


func (this *storage) Get(key string) string {
	this.RLock()
	result := this.get(key)
	this.RUnlock()
	return result
}

func (this *storage) get(key string) string {
	return this.items[key]
}

func (this *storage) Set(key string, value string) bool {
	this.Lock()
	result := this.set(key, value)
	this.Unlock()
	return result
}

func (this *storage) set(key string, value string) bool {
	this.items[key]=value
	return true
}

func (this *storage) Items() map[string]string {
	this.RLock()
	defer this.RUnlock()
	return this.items
}


func (this *storage) Persist(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("Error registering item types with Gob library")
		}
	}()
	log.Info("Saving ....")
	this.RLock()
	defer this.RUnlock()
	for _, v := range this.items {
		gob.Register(v)
	}
	err = enc.Encode(&this.items)
	return
}

func (this  *storage) PersistFile(fname string) error {
	fp, err := os.Create(fname)

	if err != nil {
		return err
	}
	err = this.Persist(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}


func (this *storage) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	f_items := map[string]string{}
	err := dec.Decode(&f_items)
	if err == nil {
		this.Lock()
		defer this.Unlock()
		this.items = f_items
	}
	return err
}


func (this *storage) LoadFile(fname string) error {
	fp, err := os.Open(fname)
	if err != nil {
		return err
	}
	err = this.Load(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}


func Init(li time.Duration) *Storage {
	s := &Storage{
		&storage{items: make(map[string]string), },
	}
	if li > 0 {
		runLoader(s.storage, li)
		runtime.SetFinalizer(s, stopLoader)
	}

	s.LoadFile("data.io")
	return s
}

type loader struct {
	Interval time.Duration
	stop     chan bool
}

func (this *loader) Run(s *storage) {
	this.stop = make(chan bool)
	tick := time.Tick(this.Interval)
	for {
		select {
		case <-tick:
			s.PersistFile("data.io")
		case <-this.stop:
			return
		}
	}
}

func stopLoader(s *Storage) {
	s.loader.stop <- true
}

func runLoader(s *storage, ci time.Duration) {
	l := &loader{
		Interval: ci,
	}
	s.loader = l
	go l.Run(s)
}

