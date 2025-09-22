package crawler

import (
	"log"
	"net/url"
	"sync"

	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_16/internal/downloader"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_16/internal/parser"
	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_16/internal/saver"
)

type Crawler struct {
	startURL     *url.URL
	maxDepth     int
	concurrency  int
	downloader   *downloader.Downloader
	parser       *parser.Parser
	saver        *saver.Saver
	visited      map[string]bool
	visitedMutex sync.Mutex
	wg           sync.WaitGroup
	queue        chan crawlRequest
}

type crawlRequest struct {
	url   *url.URL
	depth int
}

func NewCrawler(startURL *url.URL, maxDepth, concurrency int, d *downloader.Downloader, p *parser.Parser, s *saver.Saver) *Crawler {
	return &Crawler{
		startURL:    startURL,
		maxDepth:    maxDepth,
		concurrency: concurrency,
		downloader:  d,
		parser:      p,
		saver:       s,
		visited:     make(map[string]bool),
		queue:       make(chan crawlRequest, concurrency),
	}
}

func (c *Crawler) Crawl() {
	for i := 0; i < c.concurrency; i++ {
		go c.worker()
	}

	c.wg.Add(1)
	c.queue <- crawlRequest{url: c.startURL, depth: 0}

	c.wg.Wait()
	close(c.queue)
}

func (c *Crawler) worker() {
	for req := range c.queue {
		if req.depth > c.maxDepth {
			c.wg.Done()
			continue
		}

		c.visitedMutex.Lock()
		if c.visited[req.url.String()] {
			c.visitedMutex.Unlock()
			c.wg.Done()
			continue
		}
		c.visited[req.url.String()] = true
		c.visitedMutex.Unlock()

		log.Printf("download: %s on depth %d\n", req.url.String(), req.depth)

		body, contentType, err := c.downloader.Download(req.url.String())
		if err != nil {
			log.Printf("error download %s: %v\n", req.url.String(), err)
			c.wg.Done()
			continue
		}

		if c.parser.IsHTML(contentType) {
			modifiedBody, err := c.parser.RewriteLinks(body, req.url, c.saver)
			if err != nil {
				log.Printf("error rewriting links for %s: %v\n", req.url.String(), err)
			} else {
				body = modifiedBody
			}

			links := c.parser.ExtractLinks(body, req.url)
			for _, link := range links {
				if link.Host == c.startURL.Host {
					c.wg.Add(1)
					c.queue <- crawlRequest{url: link, depth: req.depth + 1}
				}
			}
		}

		_, err = c.saver.Save(req.url, body, contentType)
		if err != nil {
			log.Printf("error save %s: %v\n", req.url.String(), err)
			c.wg.Done()
			continue
		}

		c.wg.Done()
	}
}
