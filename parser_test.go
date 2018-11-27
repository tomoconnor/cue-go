package cue

import (
	"testing"
)

type expected struct {
	Cmd    string
	Params []string
}

type test struct {
	Input  string
	Etalon expected
}

func TestParseCommand(t *testing.T) {
	var tests = []test{
		{"COMMAND",
			expected{"COMMAND",
				[]string{}}},
		{"COMMAND \t PARAM1   PARAM2\tPARAM3",
			expected{"COMMAND",
				[]string{"PARAM1", "PARAM2", "PARAM3"}}},
		{"COMMAND 'PARAM1' \"PARAM2\" 'PAR\"AM3' 'P AR  AM 4'",
			expected{"COMMAND",
				[]string{"PARAM1", "PARAM2", "PAR\"AM3", "P AR  AM 4"}}},
		{"COMMAND 'P A R A M 1' \"PA RA M2\" PA\\\"RAM\\'3",
			expected{"COMMAND",
				[]string{"P A R A M 1", "PA RA M2", "PA\"RAM'3"}}},
	}

	for _, tt := range tests {
		cmd, params, err := parseCommand(tt.Input)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if cmd != tt.Etalon.Cmd {
			t.Fatalf("parsed command '%s' but '%s' expected", cmd, tt.Etalon.Cmd)
		}

		if len(params) != len(tt.Etalon.Params) {
			t.Fatalf("parsed %d params but %d expected", len(params), len(tt.Etalon.Params))
		}

		for i := 0; i < len(params); i++ {
			if params[i] != tt.Etalon.Params[i] {
				t.Fatalf("parsed '%s' parameter but '%s' expected", params[i], tt.Etalon.Params[i])
			}
		}
	}
}

type timeExpected struct {
	min    int
	sec    int
	frames int
}

func TestParseTime(t *testing.T) {
	var tests = map[string]timeExpected{
		"01:02:03": {1, 2, 3},
		"11:22:33": {11, 22, 33},
		"14:00:00": {14, 0, 0},
	}

	for input, expected := range tests {
		min, sec, frames, err := parseTime(input)
		if err != nil {
			t.Fatalf("time parsing failed, input string: '%s', error: %v", input, err)
		}

		if min != expected.min {
			t.Fatalf("expected %d minutes, but %d recieved.", expected.min, min)
		}
		if sec != expected.sec {
			t.Fatalf("expected %d seconds, but %d recieved.", expected.sec, sec)
		}
		if frames != expected.frames {
			t.Fatalf("expected %d frames, but %d recieved.", expected.frames, frames)
		}
	}
}
