package diagnostics

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/google/uuid"
)

// GetProf returns a base64 encoded png of a go cpu profile
// it takes the number of seconds to profile and the url of the profile endpoint
// (usually http://localhost:6060)
func getProf(seconds int, profileUrl string) string {
	u := uuid.New()
	fileName := fmt.Sprintf("/tmp/profile-%s.png", u.String())
	url := fmt.Sprintf("%s/debug/pprof/profile?seconds=%d", profileUrl, seconds)

	cmd := exec.Command("go", "tool", "pprof", "-output", fileName, "-png", url)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		fmt.Println("Stderr:", stderr.String())
		return ""
	}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	// todo remove file

	encoded := base64.StdEncoding.EncodeToString(data)

	return encoded
}
