package screenshotworker

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/santrancisco/chromedp"
	"github.com/santrancisco/chromedp/runner"
)

type Worker struct {
	wp  *WorkerPool
	id  int
	ctx context.Context
	URL chan string
}

func NewWorker(ctx context.Context, wp *WorkerPool, id int, url chan string) *Worker {
	return &Worker{
		id:  id,
		wp:  wp,
		ctx: ctx,
		URL: url,
	}
}

// WorkerPool is a pool of Screenshot Workers
type WorkerPool struct {
	ctx        context.Context
	workers    []*Worker
	mu         sync.Mutex
	Wg         sync.WaitGroup
	Chromepool *chromedp.Pool
	URL        chan string
	done       bool
}

func NewWorkerPool(ctx context.Context, count int) *WorkerPool {
	pool, err := chromedp.NewPool( /*chromedp.PoolLog(log.Printf, log.Printf, log.Printf)*/ )
	if err != nil {
		log.Fatal(err.Error())
	}
	return &WorkerPool{
		ctx:        ctx,
		Chromepool: pool,
		URL:        make(chan string, 200),
		workers:    make([]*Worker, count),
	}
}

// Start starts all of the Workers in the WorkerPool.
func (wp *WorkerPool) Start() {
	log.Print("[DEBUG] Starting worker nodes")
	wp.mu.Lock()
	defer wp.mu.Unlock()
	for i, _ := range wp.workers {
		wp.workers[i] = NewWorker(wp.ctx, wp, i, wp.URL)
		wp.Wg.Add(1)
		go wp.workers[i].Run()
	}
}

// Start starts all of the Workers in the WorkerPool.
func (w *Worker) Run() {
	defer w.wp.Wg.Done()
	log.Printf("[DEBUG] Starting worker node %d", w.id)

	for {
		// Using select for non-blocking reading from channels.
		select {
		case <-w.ctx.Done():
			return
		case url := <-w.URL:
			log.Printf("[DEBUG] Taking screenshot for %s\n", url)
			w.Takescreenshot(url)
		}
	}
}

func (w Worker) Takescreenshot(url string) {
	// run task list
	var buf []byte
	c, err := w.wp.Chromepool.Allocate(w.ctx,
		runner.Flag("headless", true),
		runner.Flag("ignore-certificate-errors", true),
		runner.Flag("disable-gpu", true),
		runner.WindowSize(1280, 800),
	)
	if err != nil {
		log.Printf("[DEBUG] Error: %v", err)
		return
	}
	defer c.Release()

	err = c.Run(w.ctx, screenshot(url, &buf))
	if err != nil {
		log.Fatal(err)
	}
	if len(buf) == 19809 {
		log.Printf("[DEBUG] Got a blank screenshot for %s", url)
		return
	}
	s := sha1.New()
	io.WriteString(s, url)
	filename := fmt.Sprintf("%x", s.Sum(nil))
	log.Printf("[DEBUG] Writting screenshot for %s to file %s", url, filename)
	err = ioutil.WriteFile(filename+".png", buf, 0644)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(filename+".txt", []byte(url), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func screenshot(urlstr string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Sleep(3 * time.Second),
		chromedp.CaptureScreenshot(res),
	}
}

func tagscreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Sleep(4 * time.Second),
		chromedp.WaitVisible(sel, chromedp.ByID),
		chromedp.WaitNotVisible(`div.v-middle > div.la-ball-clip-rotate`, chromedp.ByQuery),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.ByID),
	}
}
