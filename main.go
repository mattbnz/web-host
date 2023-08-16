// Github webhook handler to pull and build a Hugo website.
//
// Copyright © 2023 Matt Brown. MIT Licensed.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
)

type Config struct {
	Port string

	RepoUrl           string
	MetricsRepoUrl    string
	WorkingDir        string
	MetricsWorkingDir string
	OutputDir         string

	Valid bool
}

func (c *Config) Load() {
	c.Port = os.Getenv("PORT")
	if c.Port == "" {
		c.Port = "8080"
	}

	c.RepoUrl = os.Getenv("REPO_URL")
	c.MetricsRepoUrl = os.Getenv("METRICS_REPO_URL")
	c.WorkingDir = os.Getenv("WORKING_DIR")
	if c.WorkingDir == "" {
		c.WorkingDir = "/app/source"
	}
	c.MetricsWorkingDir = os.Getenv("METRICS_DIR")
	if c.MetricsWorkingDir == "" {
		c.MetricsWorkingDir = "/app/metrics"
	}
	c.OutputDir = os.Getenv("OUTPUT_DIR")
	if c.OutputDir == "" {
		c.OutputDir = "/app/html"
	}

	if c.RepoUrl != "" {
		c.Valid = true
	}
}

var GlobalConfig Config

// Wrapper to run a command with stdout/stderr hooked to the console
func Command(wd string, name string, args ...string) error {
	if wd == "" {
		wd = GlobalConfig.WorkingDir
	}
	fmt.Println(name, args)
	cmd := exec.Command(name, args...)
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Always returns success, since we don't want GitHub retrying.
func HandleHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusOK)
		return
	}

	// TODO: Check for actual webhook stuff...
	UpdateSite()
	w.WriteHeader(http.StatusOK)
}

func clone(repo_url, work_dir string) error {
	if _, err := os.Stat(work_dir); os.IsNotExist(err) {
		if err := Command("/app", "git", "clone", repo_url, work_dir); err != nil {
			log.Printf("Failed to clone repo: %v", err)
			return err
		}
	} else {
		if err := Command("", "git", "pull"); err != nil {
			log.Printf("Failed to pull repo: %v", err)
			return err
		}
	}
	return nil
}

func UpdateSite() {
	if err := clone(GlobalConfig.RepoUrl, GlobalConfig.WorkingDir); err != nil {
		return
	}
	if err := Command("", "npm", "ci"); err != nil {
		log.Println(err)
		return
	}
	if err := Command(path.Join(GlobalConfig.WorkingDir, "themes/default"), "npm", "ci"); err != nil {
		log.Println(err)
		return
	}
	if err := Command("", "/app/bin/hugo", "-d", GlobalConfig.OutputDir, "-e", "review", "-D", "-F"); err != nil {
		log.Println(err)
		return
	}
	log.Println("Site updated sucessfully!")
}

func main() {
	GlobalConfig.Load()
	if !GlobalConfig.Valid {
		log.Println("WARNING: invalid config - see /healthz for details!")
	}

	// Handle requests
	http.HandleFunc("/", HandleHook)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if GlobalConfig.Valid {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("all good"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			if GlobalConfig.RepoUrl == "" {
				w.Write([]byte("missing repo url\n"))
			}
		}
	})

	UpdateSite()

	log.Println("listening on", GlobalConfig.Port)
	log.Fatal(http.ListenAndServe(":"+GlobalConfig.Port, nil))
}
