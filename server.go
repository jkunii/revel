// +build !appengine

package rev

import (
	"fmt"
	"net/http"
	"path"
	"time"
)

var (
	MainRouter         *Router
	MainTemplateLoader *TemplateLoader
	MainWatcher        *Watcher
	Server             *http.Server
)

// Run the server.
// This is called from the generated main file.
// If port is non-zero, use that.  Else, read the port from app.conf.
func Run(port int) {
	address := HttpAddr
	if port == 0 {
		port = HttpPort
	}

	MainRouter = NewRouter(path.Join(BasePath, "conf", "routes"))
	MainTemplateLoader = NewTemplateLoader(TemplatePaths)

	// If desired (or by default), create a watcher for templates and routes.
	// The watcher calls Refresh() on things on the first request.
	if Config.BoolDefault("server.watcher", true) {
		MainWatcher = NewWatcher()
		MainWatcher.auditor = PluginNotifier{plugins}
		MainWatcher.Listen(MainTemplateLoader, MainTemplateLoader.paths...)
		MainWatcher.Listen(MainRouter, MainRouter.path)
	} else {
		// Else, call refresh on them directly.
		MainTemplateLoader.Refresh()
		MainRouter.Refresh()
		plugins.OnRoutesLoaded(MainRouter)
	}

	Server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", address, port),
		Handler: http.HandlerFunc(handle),
	}

	plugins.OnAppStart()

	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("Listening on port %d...\n", port)
	}()

	ERROR.Fatalln("Failed to listen:", Server.ListenAndServe())
}

// The PluginNotifier glues the watcher and the plugin collection together.
// It audits refreshes and invokes the appropriate method to inform the plugins.
type PluginNotifier struct {
	plugins PluginCollection
}

func (pn PluginNotifier) OnRefresh(l Listener) {
	if l == MainRouter {
		pn.plugins.OnRoutesLoaded(MainRouter)
	}
}
