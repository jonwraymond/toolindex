package toolindex

import (
	"errors"
	"reflect"
	"testing"

	"github.com/jonwraymond/toolmodel"
)

func TestIndexContract_ErrorSentinels(t *testing.T) {
	idx := NewInMemoryIndex()

	_, _, err := idx.GetTool("missing:tool")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetTool error = %v, want ErrNotFound", err)
	}

	err = idx.RegisterTool(toolmodel.Tool{}, toolmodel.ToolBackend{})
	if !errors.Is(err, ErrInvalidTool) {
		t.Fatalf("RegisterTool error = %v, want ErrInvalidTool", err)
	}
}

func TestSearcherContract_LexicalDeterminism(t *testing.T) {
	searcher := &lexicalSearcher{}
	ds, ok := interface{}(searcher).(DeterministicSearcher)
	if !ok || !ds.Deterministic() {
		t.Fatalf("lexicalSearcher should be deterministic")
	}

	docs := []SearchDoc{
		{ID: "a:one", DocText: "one alpha", Summary: Summary{ID: "a:one", Name: "one"}},
		{ID: "b:two", DocText: "two beta", Summary: Summary{ID: "b:two", Name: "two"}},
	}
	first, err := searcher.Search("o", 10, docs)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	second, err := searcher.Search("o", 10, docs)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("Search results should be deterministic")
	}
}

func TestSearcherContract_ZeroLimit(t *testing.T) {
	searcher := &lexicalSearcher{}
	results, err := searcher.Search("anything", 0, nil)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty results for limit 0, got %d", len(results))
	}
}

func TestChangeNotifierContract_Unsubscribe(t *testing.T) {
	idx := NewInMemoryIndex()
	unsub := idx.OnChange(nil)
	if unsub == nil {
		t.Fatalf("expected non-nil unsubscribe")
	}
	unsub()
	unsub() // should be idempotent
}

func TestRefresherContract_MonotonicVersion(t *testing.T) {
	idx := NewInMemoryIndex()
	v1 := idx.Refresh()
	v2 := idx.Refresh()
	if v2 <= v1 {
		t.Fatalf("expected monotonic version, got v1=%d v2=%d", v1, v2)
	}
}
