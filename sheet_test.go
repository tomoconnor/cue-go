package cue

import "testing"

func TestTime_Seconds(t *testing.T) {
	expect := 70.0 + 10.0/framesPerSecond
	time := Time{
		Min:    1,
		Sec:    10,
		Frames: 10,
	}
	s := time.Seconds()
	if s != expect {
		t.Errorf("expect len: %f, got: %f", expect, s)
	}
}
