package diagnostics

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/google/uuid"
)

// GetProf returns a base64 encoded png of a go cpu profile
// To not requite golang installed on the machine we bundle pprof and call the
// pprof driver directly.
func getProf(profileUrl string) string {
	u := uuid.New()
	fileName := fmt.Sprintf("/tmp/profile-%s.png", u.String())

	current, err := os.Executable()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(current, "profile", "-o", fileName, "-p", profileUrl)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		fmt.Println("Stderr:", stderr.String())
		return ""
	}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	// remove file
	err = os.Remove(fileName)
	if err != nil {
		panic(err)
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	return encoded
}
