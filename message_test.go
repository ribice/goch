package goch_test

import (
	"reflect"
	"testing"

	"github.com/ribice/goch"
)

func TestEncodeMSG(t *testing.T) {
	m := &goch.Message{Time: 123, Seq: 1, Text: "Hello World", FromUID: "ABC", FromName: "User1"}
	bts, err := m.Encode()
	if err != nil {
		t.Errorf("did not expect error but received: %v", err)
	}
	msg, err := goch.DecodeMsg(bts)
	if err != nil {
		t.Errorf("did not expect error but received: %v", err)
	}
	if !reflect.DeepEqual(m, msg) {
		t.Errorf("expected msg %v but got %v", m, msg)
	}

	_, err = goch.DecodeMsg([]byte("msg"))
	if err == nil {
		t.Error("expected error but received nil")
	}
}
