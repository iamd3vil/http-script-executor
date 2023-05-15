package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
	scriptsPath := path.Clean(ko.String("script_folder"))
	g.POST(`/{script:*}`, handleExecuteScript(scriptsPath))

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

		scriptPath := path.Join(scripts, script)
		fmt.Printf("script: %s\n", scriptPath)
		// Check if the base dir of script path is the same as script dir.
		if !isSubPath(scripts, scriptPath) {
			r.RequestCtx.SetStatusCode(fasthttp.StatusNotFound)
			r.RequestCtx.WriteString("error finding given script")
			return nil
		}

		// Check if the given script is in the scripts directory.
		if _, err := os.Stat(scriptPath); err != nil {
			r.RequestCtx.SetStatusCode(fasthttp.StatusNotFound)
			r.RequestCtx.WriteString("error finding given script")
			return nil
		}

		log.Printf("executing: %s %s", scriptPath, strings.Join(args, " "))
		output, err := exec.Command(scriptPath, args...).CombinedOutput()
		if err != nil {
			log.Printf("error executing, err: %s, output: %s", err.Error(), output)
			r.RequestCtx.WriteString(err.Error())
			r.RequestCtx.SetStatusCode(fasthttp.StatusInternalServerError)
			return nil
		}
		log.Println(string(output))
		log.Printf("finished executing: %s %s", scriptPath, strings.Join(args, " "))

		fmt.Fprint(r.RequestCtx, string(output))
		return nil
	}
}

func isSubPath(parentPath, childPath string) bool {
	parentPath = path.Clean(parentPath)
	childPath = path.Clean(childPath)

	absParent, err := filepath.Abs(parentPath)
	if err != nil {
		return false
	}

	absChild, err := filepath.Abs(childPath)
	if err != nil {
		return false
	}

	return strings.HasPrefix(absChild, absParent)
}
