package dingtalk

import "testing"

func TestSignRequestData(t *testing.T) {

	r := SignRequestData(1593616850977, "DOUBLEMINEKUBEWATCH")
	t.Log("sign result:" + r)
	if r != "vBL8wcjoonhfSobGT/97jcxKXxc69RW4bXTquF6HCyQ=" {
		t.Error("the result not match")
	}

}
