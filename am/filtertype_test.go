package am_test

import (
	"testing"

	"github.com/linkai-io/am/am"
)

func TestFilterBools(t *testing.T) {
	f := &am.FilterType{}
	f.AddBool("test", true)
	val, ok := f.Bool("test")
	if !ok || !val {
		t.Fatalf("did not return proper value")
	}
	f.AddBool("test", false)

	vals, _ := f.Bools("test")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d\n", len(vals))
	}

	if vals[0] != true || vals[1] != false {
		t.Fatalf("expected true false, got %v %v\n", vals[0], vals[1])
	}

	f = &am.FilterType{}
	f.AddBools("test", []bool{true, false})
	vals, _ = f.Bools("test")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d\n", len(vals))
	}

	if vals[0] != true || vals[1] != false {
		t.Fatalf("expected true false, got %v %v\n", vals[0], vals[1])
	}
}

func TestFilterInt32s(t *testing.T) {
	f := &am.FilterType{}
	f.AddInt32("test", 32)
	val, ok := f.Int32("test")
	if !ok || val != 32 {
		t.Fatalf("did not return proper value")
	}
	f.AddInt32("test", 33)

	vals, _ := f.Int32s("test")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d\n", len(vals))
	}

	if vals[0] != 32 || vals[1] != 33 {
		t.Fatalf("expected 32 33, got %v %v\n", vals[0], vals[1])
	}

	f = &am.FilterType{}
	f.AddInt32s("test", []int32{32, 33})

	vals, _ = f.Int32s("test")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d\n", len(vals))
	}

	if vals[0] != 32 || vals[1] != 33 {
		t.Fatalf("expected 32 33, got %v %v\n", vals[0], vals[1])
	}
}

func TestFilterInt64s(t *testing.T) {
	f := &am.FilterType{}
	f.AddInt64("test", 3200000000000)
	val, ok := f.Int64("test")
	if !ok || val != 3200000000000 {
		t.Fatalf("did not return proper value")
	}
	f.AddInt64("test", 3300000000000)

	vals, _ := f.Int64s("test")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d\n", len(vals))
	}

	if vals[0] != 3200000000000 || vals[1] != 3300000000000 {
		t.Fatalf("expected 3200000000000 3300000000000, got %v %v\n", vals[0], vals[1])
	}

	f = &am.FilterType{}
	f.AddInt64s("test", []int64{3200000000000, 3300000000000})

	vals, _ = f.Int64s("test")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d\n", len(vals))
	}

	if vals[0] != 3200000000000 || vals[1] != 3300000000000 {
		t.Fatalf("expected 3200000000000 3300000000000, got %v %v\n", vals[0], vals[1])
	}
}

func TestFilterFloat32s(t *testing.T) {
	f := &am.FilterType{}
	f.AddFloat32("test", 32.2)
	val, ok := f.Float32("test")
	if !ok || val != 32.2 {
		t.Fatalf("did not return proper value")
	}
	f.AddFloat32("test", 33.3)

	vals, _ := f.Float32s("test")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d\n", len(vals))
	}

	if vals[0] != 32.2 || vals[1] != 33.3 {
		t.Fatalf("expected 32.2 33.3, got %v %v\n", vals[0], vals[1])
	}

	f = &am.FilterType{}
	f.AddFloat32s("test", []float32{32.2, 33.3})

	vals, _ = f.Float32s("test")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d\n", len(vals))
	}

	if vals[0] != 32.2 || vals[1] != 33.3 {
		t.Fatalf("expected 32.2 33.3, got %v %v\n", vals[0], vals[1])
	}
}

func TestFilterStrings(t *testing.T) {
	f := &am.FilterType{}
	f.AddString("test", "32")
	val, ok := f.String("test")
	if !ok || val != "32" {
		t.Fatalf("did not return proper value")
	}
	f.AddString("test", "33")

	vals, _ := f.Strings("test")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d\n", len(vals))
	}

	if vals[0] != "32" || vals[1] != "33" {
		t.Fatalf("expected 32 33, got %v %v\n", vals[0], vals[1])
	}

	f = &am.FilterType{}
	f.AddStrings("test", []string{"32", "33"})

	vals, _ = f.Strings("test")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d\n", len(vals))
	}

	if vals[0] != "32" || vals[1] != "33" {
		t.Fatalf("expected 32 33, got %v %v\n", vals[0], vals[1])
	}
}
