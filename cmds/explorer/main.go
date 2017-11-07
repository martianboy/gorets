package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/jpfielding/gorets/config"
	"github.com/jpfielding/gorets/explorer"
)

func main() {
	port := flag.String("port", "8000", "http port")
	react := flag.String("react", "../../explorer/client/build", "ReactJS path")
	configPath := flag.String("configs", "source-configs", "The configurations for this service")

	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(*react)))

	// HACK were moving to loading from a web endpoint anyways...
	cfgs, _ := config.ImportFrom(*configPath)
	fmt.Printf("loaded %d configs\n", len(cfgs))

	// gorilla rpc
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(&explorer.ConnectionService{Connections: cfgs}, "")
	s.RegisterService(&explorer.MetadataService{}, "")
	s.RegisterService(&explorer.SearchService{}, "")
	s.RegisterService(&explorer.ObjectService{}, "")

	cors := handlers.CORS(
		handlers.AllowedMethods([]string{"OPTIONS", "POST", "GET", "HEAD"}),
		handlers.AllowedHeaders([]string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"}),
	)
	// rpc calls
	http.Handle("/rpc", handlers.CompressHandler(cors(s)))

	// websocket wire logs
	http.Handle("/wirelog", explorer.WireLogSocket(cfgs, explorer.WirelogUpgrader))

	log.Println("Server starting: http://localhost:" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
