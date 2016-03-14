package compose

import (
	"bytes"
	"os"
	"testing"
)

var goodYML = `
test_mockserver:
  container_name: ms
  image: jamesdbloom/mockserver
  ports:
    - "10000:1080"
    - "1090"
test_postgres:
  container_name: pg
  image: postgres
  ports:
    - "5432"
`

var badYML = `
bad
`

func TestGoodYML(t *testing.T) {
	compose := MustStart(goodYML, true, true)
	defer compose.MustKill()
	if compose.Containers["ms"].Name != "/ms" {
		t.Fatalf("found name '%v', expected '/ms", compose.Containers["ms"].Name)
	}
	if compose.Containers["pg"].Name != "/pg" {
		t.Fatalf("found name '%v', expected '/pg", compose.Containers["pg"].Name)
	}
	if port := compose.Containers["ms"].MustGetFirstPublicPort(1080, "tcp"); port != 10000 {
		t.Fatalf("found port %v, expected 10000", port)
	}

}

func TestBadYML(t *testing.T) {
	compose, err := Start(badYML, true, true)
	if err == nil {
		defer compose.MustKill()
		t.FailNow()
	}
}

func TestInferDockerHost(t *testing.T) {
	envHost := os.Getenv("DOCKER_HOST")
	defer os.Setenv("DOCKER_HOST", envHost)

	os.Setenv("DOCKER_HOST", "")
	if host, _ := InferDockerHost(); host != "localhost" {
		t.Errorf("found '%v', expected 'localhost'", host)
	}
	os.Setenv("DOCKER_HOST", "tcp://192.168.99.100:2376")
	if host, _ := InferDockerHost(); host != "192.168.99.100" {
		t.Errorf("found '%v', expected '192.168.99.100'", host)
	}
	os.Setenv("DOCKER_HOST", "bad")
	if _, err := InferDockerHost(); err == nil {
		t.Fail()
	}
}

func TestMultiWrite(t *testing.T) {
	var w1, w2 bytes.Buffer
	mw := newMultiWriter(&w1, &w2)
	n, err := mw.Write([]byte("test"))
	if err != nil || n != len("test") {
		t.Fatalf("expected no error, got %v, %v", err, n)
	}
	if w1.String() != "test" {
		t.Fatal("output not piped correctly to w1")
	}
	if w2.String() != "test" {
		t.Fatal("output not piped correctly to w2")
	}
}
