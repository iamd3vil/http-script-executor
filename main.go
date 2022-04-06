package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

func initConfig(path string) (*koanf.Koanf, error) {
	ko := koanf.New(".")
	if err := ko.Load(file.Provider(path), toml.Parser()); err != nil {
		return nil, err
	}

	return ko, nil
}

func main() {
	cfgPath := flag.String("config", "config.toml", "Configuration path.")
	flag.Parse()

	ko, err := initConfig(*cfgPath)
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	g := fastglue.New()
	g.GET("/", handleIndex())
	g.POST("/{script}", handleExecuteScript(ko.String("script_folder")))

	s := &fasthttp.Server{
		Name:         "HTTP Script Executor",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s, script directory: %s", ko.String("addr"), ko.String("script_folder"))

	if err := g.ListenAndServe(ko.String("addr"), "", s); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func handleIndex() fastglue.FastRequestHandler {
	return func(r *fastglue.Request) error {
		r.RequestCtx.WriteString("Hello from http-script-executor.")
		r.RequestCtx.SetStatusCode(fasthttp.StatusOK)
		return nil
	}
}

func handleExecuteScript(scripts string) fastglue.FastRequestHandler {
	return func(r *fastglue.Request) error {
		script := r.RequestCtx.UserValue("script").(string)
		args := []string{}
		json.Unmarshal(r.RequestCtx.PostBody(), &args)

		// Check if the given script is in the scripts directory.
		if _, err := os.Stat(path.Join(scripts, script)); err != nil {
			r.RequestCtx.SetStatusCode(fasthttp.StatusNotFound)
			r.RequestCtx.WriteString("error finding given script")
			return nil
		}

		log.Printf("executing: %s %s", path.Join(scripts, script), strings.Join(args, " "))
		output, err := exec.Command(path.Join(scripts, script), args...).CombinedOutput()
		if err != nil {
			r.RequestCtx.WriteString(err.Error())
			r.RequestCtx.SetStatusCode(fasthttp.StatusInternalServerError)
			return nil
		}

		fmt.Fprint(r.RequestCtx, string(output))
		return nil
	}
}
