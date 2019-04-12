package errid

import (
	"log"
	"testing"
)

func TestFormat(t *testing.T) {
	err := New("hello")
	log.Printf("err1=%+v\n", err)
}
