package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/sarcb/catalyst-sp24"
	"github.com/sarcb/catalyst-sp24/cmd"
	"github.com/sarcb/catalyst-sp24/generated/api"
	"github.com/sarcb/catalyst-sp24/hooks"
	"github.com/sarcb/catalyst-sp24/ui"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config, err := cmd.ParseCatalystConfig()
	if err != nil {
		log.Fatal(err)
	}

	theCatalyst, err := catalyst.New(&hooks.Hooks{}, config)
	if err != nil {
		log.Fatal(err)
	}

	fsys, _ := fs.Sub(ui.UI, "dist")
	staticHandlerFunc := http.HandlerFunc(api.VueStatic(fsys))
	theCatalyst.Server.Get("/ui/*", http.StripPrefix("/ui", staticHandlerFunc).ServeHTTP)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.Port),
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           theCatalyst.Server,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
