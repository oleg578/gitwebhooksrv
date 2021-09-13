package main

import (
	"github.com/google/go-github/v39/github"
	logger "github.com/oleg578/loglog"
	"net/http"
	"os"
	"os/exec"
)

func payloadHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	//validate
	payload, err := github.ValidatePayload(r, []byte("bwretail-fixmytoys"))
	if err != nil {
		logger.Printf("payload is not valid: %v", err)
		return
	}
	logger.Printf("payload length: %v", len(payload))
	if err := os.Chdir("/srv/icomdc.com"); err != nil {
		logger.Print(err)
	}
	cmd := exec.Command("git", "pull")
	if err := cmd.Run(); err != nil {
		logger.Print(err)
	}

}
