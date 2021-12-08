package skiplist

import (
	"testing"
)

func buildSkiplist() *Skiplist {
	sl := New()

	sl.Put("000", "000")
	sl.Put("001", "001")
	sl.Put("008", "008")
	sl.Put("003", "003")
	sl.Put("005", "005")
	sl.Put("007", "007")
	sl.Put("006", "006")
	sl.Put("004", "004")
	sl.Put("004", "004-rewrite")

	return sl
}

func TestLength(t *testing.T) {
	sl := buildSkiplist()
	l := sl.Length()

	if l != 8 {
		t.Errorf("Length() is %d, should be 8", l)
	}
}

func TestGet(t *testing.T) {
	sl := buildSkiplist()

	v, err := sl.Get("004")

	if err != nil || v != "004-rewrite" {
		t.Errorf("Get(004) is %s and error is %v, should be '004-rewrite'", v, err)
	}

	_, err = sl.Get("009")
	if err == nil {
		t.Errorf("Get(009) should return error(Not Found)")
	}
}

func TestDel(t *testing.T) {
	sl := buildSkiplist()

	err := sl.Del("009")
	if err == nil {
		t.Errorf("Del(009) should return error(Not Found)")
	}

	err = sl.Del("005")
	if err != nil {
		t.Errorf("Del(005) should return nil error")
	}

	err = sl.Del("005")
	if err == nil {
		t.Errorf("Del(005) should return error(Not Found)")
	}

	_, err = sl.Get("005")
	if err == nil {
		t.Errorf("Get(005) should return error(Not Found)")
	}
}

func TestRange(t *testing.T) {
	sl := buildSkiplist()

	r, _ := sl.RangeByKey("001", "003")
	if len(r) != 2 || r["001"] != "001" || r["003"] != "003" {
		t.Errorf("RangeByKey error, the result=%v", r)
	}

	r, _ = sl.RangeByCount("002", 2)
	if len(r) != 2 || r["003"] != "003" || r["004"] != "004-rewrite" {
		t.Errorf("RangeByCount(002, 2) error, the result=%v", r)
	}

	r, _ = sl.RangeByCount("008", 2)
	if len(r) != 1 || r["008"] != "008" {
		t.Errorf("RangeByCount(008, 2) error, the result=%v", r)
	}

	r, _ = sl.RangeByCount("002", -2)
	if len(r) != 2 || r["000"] != "000" || r["001"] != "001" {
		t.Errorf("RangeByCount(002, -2), the result=%v", r)
	}

	r, _ = sl.RangeByCount("008", -2)
	if len(r) != 2 || r["008"] != "008" || r["007"] != "007" {
		t.Errorf("RangeByCount(008, -2) error, the result=%v", r)
	}

	r, _ = sl.RangeByCount("008", -2)
	if len(r) != 2 || r["008"] != "008" || r["007"] != "007" {
		t.Errorf("RangeByCount(009, -2) error, the result=%v", r)
	}

	r, _ = sl.RangeByIndex(3, 2)
	if len(r) != 2 || r["004"] != "004-rewrite" || r["005"] != "005" {
		t.Errorf("RangeByIndex(3, 2), the result=%v", r)
	}

	r, _ = sl.RangeByIndex(-1, 2)
	if len(r) != 1 || r["008"] != "008" {
		t.Errorf("RangeByIndex(-1, 2), the result=%v", r)
	}

	r, _ = sl.RangeByIndex(-2, 2)
	if len(r) != 2 || r["008"] != "008" || r["007"] != "007" {
		t.Errorf("RangeByIndex(-2, 2), the result=%v", r)
	}

	r, _ = sl.RangeByIndex(-3, 2)
	if len(r) != 2 || r["006"] != "006" || r["007"] != "007" {
		t.Errorf("RangeByIndex(-3, 2), the result=%v", r)
	}

	r, _ = sl.RangeByIndex(-1, -2)
	if len(r) != 2 || r["008"] != "008" || r["007"] != "007" {
		t.Errorf("RangeByIndex(-1, -2), the result=%v", r)
	}

	r, _ = sl.RangeByIndex(-2, -2)
	if len(r) != 2 || r["006"] != "006" || r["007"] != "007" {
		t.Errorf("RangeByIndex(-2, -2), the result=%v", r)
	}

	r, _ = sl.RangeByIndex(-5, -5)
	if len(r) != 4 || r["004"] != "004-rewrite" || r["003"] != "003" || r["001"] != "001" || r["000"] != "000" {
		t.Errorf("RangeByIndex(-5, 5), the result=%v", r)
	}
}
