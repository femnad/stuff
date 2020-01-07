package main

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const (
	tempFile = "/tmp/rabn/test.yml"
	barCount = 12
	fooCount = 6
	bazKey = "baz"
	bazCount = 6
	quxCount = 1
)

func TestSerialization(t *testing.T) {
	h := make(history)
	h["bar"] = 1
	err := h.serialize(tempFile)
	if err != nil {
		t.Errorf("failage: %s", err)
	}
	err = os.Remove(tempFile)
	if err != nil {
		t.Errorf("failage: %s", err)
	}
}


func testFail(t *testing.T, err error, msg string, args ...interface{}) {
	if err != nil {
		msg = fmt.Sprintf(msg, args)
		t.Errorf("Test fail: %s %s", err, msg)
	}
}

func TestDeserialization(t *testing.T) {
	inputMap := map[string]int{"bar": barCount, "foo": fooCount, "baz": bazCount, "qux": quxCount}
	input, err := yaml.Marshal(inputMap)
	testFail(t, err, "Marshalling")
	dir, _ := path.Split(tempFile)
	err = os.MkdirAll(dir, 0700)
	testFail(t, err, "Mkdirall")
	err = ioutil.WriteFile(tempFile, input, 0600)
	testFail(t, err, "Writing file")
	h := history{}
	err = h.deserialize(tempFile)
	testFail(t, err, "Deserializing")
	count, _ := h[bazKey]
	if count != bazCount {
		t.Errorf("failage: count for %s = %d, expected %d", bazKey, count, bazCount)
	}

	err = os.Remove(tempFile)
	testFail(t, err, "removing test file")
}

func TestUpdate(t *testing.T) {
	inputMap := map[string]int{"bar": barCount, "foo": fooCount, "baz": bazCount, "qux": quxCount}
	input, err := yaml.Marshal(inputMap)
	testFail(t, err, "Marshalling")
	dir, _ := path.Split(tempFile)
	err = os.MkdirAll(dir, 0700)
	testFail(t, err, "Mkdirall")
	err = ioutil.WriteFile(tempFile, input, 0600)
	testFail(t, err, "Writing file")
	h := history{}
	err = h.deserialize(tempFile)
	testFail(t, err, "Deserializing")

	addToHistory(bazKey, tempFile)
	testFail(t, err, "Adding to history")

	err = h.deserialize(tempFile)
	testFail(t, err, "Deserializing")
	newCount, _ := h[bazKey]
	expectedCount := bazCount + 1
	if newCount != expectedCount {
		t.Errorf("failage: count for %s = %d, expected %d", bazKey, newCount, expectedCount)
	}

	err = os.Remove(tempFile)
	testFail(t, err, "removing test file")
}
