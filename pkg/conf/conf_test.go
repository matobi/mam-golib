package conf

import (
	"math/rand"
	"strconv"
	"testing"
)

const (
	confPort = "port"
)

func TestEmpty(t *testing.T) {
	c := NewConfig("test")
	_, err := c.LogAndValidate()
	if err != nil {
		t.Errorf("empty cfg not validated")
	}
}

func TestAddGet(t *testing.T) {
	data := []struct {
		t     ValueType
		name  string
		value string
	}{
		{VtStr, "str1", "str1"},
		{VtStr, "str2", "str2"},
		{VtInt, "int1", "1"},
		{VtInt, "int2", "123456789"},
		{VtInt, "int3", "-12345"},
	}
	c := NewConfig("test")
	for _, d := range data {
		c.Add(d.t, d.name, d.value)
	}

	_, err := c.LogAndValidate()
	if err != nil {
		t.Errorf("unexpected failed validate")
	}

	for _, d := range data {
		value := c.Str(d.name)
		if d.value != value {
			t.Errorf("unexpected value; n=%s; v=%s; got=%s", d.name, d.value, value)
		}
		if d.t == VtInt && toInt(d.value) != toInt(value) {
			t.Errorf("unexpected int value; n=%s; v=%d; got=%d", d.name, toInt(d.value), toInt(value))
		}
	}
}

func toInt(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1000000 - rand.Int63n(1000000)
	}
	return n
}

func TestOverwrite(t *testing.T) {
	name := "str1"
	value1 := "value1"
	value2 := "value2"
	c := NewConfig("")

	c.Add(VtStr, name, value1)
	v := c.Str(name)
	if v != value1 {
		t.Errorf("unexpected value; n=%s; v=%s; got=%s", name, value1, v)
	}

	c.Add(VtStr, name, value2)
	v = c.Str(name)
	if v != value2 {
		t.Errorf("unexpected value; n=%s; v=%s; got=%s", name, value2, v)
	}
}

func TestProfile(t *testing.T) {
	c := NewConfig("profActive")
	c.Add(VtStr, "nameAll", "valueAll")
	c.AddProfile(VtStr, "profActive", "name", "valueActive")
	c.AddProfile(VtStr, "profInctive", "name", "valueInactive")

	if "valueAll" != c.Str("nameAll") {
		t.Errorf("value for global profile missing; exp=%s; got=%s", "valueAll", c.Str("nameAll"))
	}
	if "valueActive" != c.Str("name") {
		t.Errorf("wrong value for active profile; exp=%s; got=%s", "valueActive", c.Str("name"))
	}
}

func TestBadInt(t *testing.T) {
	c := NewConfig("")
	c.Add(VtInt, "int", "hello")
	_, err := c.LogAndValidate()
	if err == nil {
		t.Errorf("missing validate error")
	}
}
