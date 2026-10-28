package shareddomain

import "testing"

func TestNewPagination(t *testing.T) {
	p := NewPagination(0, -1)
	if p.Limit != DefaultLimit || p.Offset != 0 {
		t.Fatalf("got %+v", p)
	}
	p = NewPagination(200, 5)
	if p.Limit != MaxLimit || p.Offset != 5 {
		t.Fatalf("got %+v", p)
	}
	p = NewPagination(10, 20)
	if p.Limit != 10 || p.Offset != 20 {
		t.Fatalf("got %+v", p)
	}
}

func TestClamp(t *testing.T) {
	l, o := Clamp(-1, -2)
	if l != DefaultLimit || o != 0 {
		t.Fatalf("got %d %d", l, o)
	}
	l, o = Clamp(50, 10)
	if l != 50 || o != 10 {
		t.Fatalf("got %d %d", l, o)
	}
}
