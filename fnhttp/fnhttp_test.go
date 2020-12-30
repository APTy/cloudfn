package fnhttp

import (
	"fmt"
	"testing"
)

type foo struct{}

type testReq struct {
	Foo *foo `json:"foo" cloudfn:"required"`
}

func TestCheckRequiredFields(t *testing.T) {
	req := &testReq{Foo: &foo{}}
	err := checkRequiredFields(req)
	fmt.Println("err:", err)
	t.Fail()
}
