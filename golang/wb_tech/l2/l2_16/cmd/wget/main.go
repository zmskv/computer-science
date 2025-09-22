package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_16/internal/crawler"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_16/internal/downloader"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_16/internal/parser"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_16/internal/saver"
)

func main() {
	targetURL := flag.String("url", "", "target url")
	depth := flag.Int("depth", 1, "depth recursion")
	flag.Parse()

	if *targetURL == "" {
		log.Fatal("must have targer url -url")
	}

	u, err := url.Parse(*targetURL)
	if err != nil {
		log.Fatalf("invalid URL: %v", err)
	}

	outputDir := u.Hostname()
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("error create directory: %v", err)
	}

	s, err := saver.NewSaver(outputDir)
	if err != nil {
		log.Fatalf("error create saver: %v", err)
	}

	d := downloader.NewDownloader(10 * time.Second)
	p := parser.NewParser()

	c := crawler.NewCrawler(u, *depth, runtime.NumCPU(), d, p, s)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Crawl()
	}()

	wg.Wait()
	fmt.Println("mirroring done")
}
