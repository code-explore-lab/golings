package catalog

import (
	"testing"
	"testing/fstest"
)

func TestLoadSortsAndValidatesExercises(t *testing.T) {
	f := fstest.MapFS{"z/exercise.json": {Data: []byte("{\"id\":\"z\",\"title\":\"Z\",\"check\":\"go-test\"}")}, "a/exercise.json": {Data: []byte("{\"id\":\"a\",\"title\":\"A\",\"check\":\"go-test\"}")}}
	got, err := Load(f)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].ID != "a" || got[1].ID != "z" {
		t.Fatalf("not sorted: %#v", got)
	}
}
func TestLoadRejectsUnsupportedChecks(t *testing.T) {
	f := fstest.MapFS{"x/exercise.json": {Data: []byte("{\"id\":\"x\",\"title\":\"X\",\"check\":\"shell\"}")}}
	if _, err := Load(f); err == nil {
		t.Fatal("expected unsupported check to fail")
	}
}
