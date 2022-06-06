package main

import (
	"crypto/tls"
	"github.com/ghostljj/gspider"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
)

func main() {

	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)

	gin.SetMode(gin.DebugMode)

	router := gin.New()

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "找不到该路由",
		})
		return
	})

	router.GET("/", indexHandler)
	srv := &http.Server{
		Addr:    ":444",
		Handler: router,
	}
	//不加这段，就是单向验证 TLS
	//加这段，就是双向验证 mTLS
	srv.TLSConfig = &tls.Config{
		ClientCAs:  gspider.LoadCaFile(currentDir + "/x509/c.ca"),
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	go func() {
		//if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		if err := srv.ListenAndServeTLS(currentDir+"/x509/s.crt", currentDir+"/x509/s.key"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	select {}
}
func indexHandler(c *gin.Context) {
	c.String(http.StatusOK, "hello world")
	return
}
