package server

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewHandler(gatherer prometheus.Gatherer) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	_ = router.SetTrustedProxies(nil)

	router.GET("/", rootHandler)
	router.GET("/healthz", healthHandler)
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})))

	registerPprof(router)

	return router
}

func rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":      "go-analyse",
		"status":    "ok",
		"metrics":   "/metrics",
		"healthz":   "/healthz",
		"pprof":     "/debug/pprof/",
		"framework": "gin",
	})
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func registerPprof(router *gin.Engine) {
	router.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	router.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	router.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	router.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	router.POST("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	router.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
	router.GET("/debug/pprof/allocs", gin.WrapH(pprof.Handler("allocs")))
	router.GET("/debug/pprof/block", gin.WrapH(pprof.Handler("block")))
	router.GET("/debug/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	router.GET("/debug/pprof/heap", gin.WrapH(pprof.Handler("heap")))
	router.GET("/debug/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))
	router.GET("/debug/pprof/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
}
