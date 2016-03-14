package compose

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"sync"
)

var dockerHostRegexp = regexp.MustCompile("://([^:]+):")

// InferDockerHost returns the current docker host based on the contents of the DOCKER_HOST environment variable.
// If DOCKER_HOST is not set, it returns "localhost".
func InferDockerHost() (string, error) {
	envHost := os.Getenv("DOCKER_HOST")
	if len(envHost) == 0 {
		return "localhost", nil
	}

	matches := dockerHostRegexp.FindAllStringSubmatch(envHost, -1)
	if len(matches) != 1 || len(matches[0]) != 2 {
		return "", fmt.Errorf("cannot parse DOCKER_HOST '%v'", envHost)
	}
	return matches[0][1], nil
}

func runCmd(name string, args ...string) (string, string, error) {
	var stdout, stderr, combined bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = newMultiWriter(&stdout, &combined)
	cmd.Stderr = newMultiWriter(&stderr, &combined)
	err := cmd.Run()
	if err != nil {
		fmt.Print(combined.String())
	}
	return stdout.String(), stderr.String(), err
}

func writeTmp(content string) (string, error) {
	f, err := ioutil.TempFile("", "docker-compose-")
	if err != nil {
		return "", fmt.Errorf("compose: error creating temp file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return "", fmt.Errorf("compose: error writing temp file: %v", err)
	}

	return f.Name(), nil
}

type multiWriter struct {
	mutex *sync.Mutex
	writers []io.Writer
}

func newMultiWriter(writers ...io.Writer) io.Writer {
	return &multiWriter{mutex: &sync.Mutex{}, writers: writers}
}

func (mw *multiWriter) Write(p []byte) (n int, err error) {
	mw.mutex.Lock()
	defer mw.mutex.Unlock()
	for i := 0; i < len(mw.writers); i++ {
		if n, err := mw.writers[i].Write(p); err != nil {
			return n, err
		}
	}
	return len(p), nil
}
