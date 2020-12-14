// Package cue implement CUE-SHEET files parser.
// For CUE documentation see: http://digitalx.org/cue-sheet/syntax/
package cue

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
)

// commandParser is the function for parsing one command.
type commandParser func(params []string, sheet *Sheet) error

// commandParserDescriptor describes command parser.
type commandParserDescriptor struct {
	// -1 -- zero or more parameters.
	paramsCount int
	parser      commandParser
}

// parsersMap used for commands and parser functions correspondence.
var parsersMap = map[string]commandParserDescriptor{
	"CATALOG":    {1, parseCatalog},
	"CDTEXTFILE": {1, parseCdTextFile},
	"FILE":       {2, parseFile},
	"FLAGS":      {-1, parseFlags},
	"INDEX":      {2, parseIndex},
	"ISRC":       {1, parseIsrc},
	"PERFORMER":  {1, parsePerformer},
	"POSTGAP":    {1, parsePostgap},
	"PREGAP":     {1, parsePregap},
	"REM":        {-1, parseRem},
	"SONGWRITER": {1, parseSongWriter},
	"TITLE":      {1, parseTitle},
	"TRACK":      {2, parseTrack},
}

// Parse parses cue-sheet data (file) and returns filled Sheet struct.
func Parse(reader io.Reader, durations ...float64) (sheet *Sheet, err error) {
	sheet = new(Sheet)

	rd := bufio.NewReader(reader)
	lineNumber := 0

	for buf, _, err := rd.ReadLine(); err != io.EOF; buf, _, err = rd.ReadLine() {
		if err != nil {
			return nil, err
		}

		line, _, _ := transform.String(runes.Remove(runes.In(unicode.Mn)), string(buf))
		line = strings.TrimSpace(line)

		// Skip empty lines.
		if len(line) == 0 {
			continue
		}

		lineNumber++

		cmd, params, err := parseCommand(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", lineNumber, err)
		}

		parserDescriptor, ok := parsersMap[cmd]
		if !ok {
			return nil, fmt.Errorf("line %d: unknown command '%s'", lineNumber, cmd)
		}

		paramsExpected := parserDescriptor.paramsCount
		paramsReceived := len(params)
		if paramsExpected != -1 && paramsExpected != paramsReceived {
			return nil, fmt.Errorf("line %d: command %s: recieved %d parameters but %d expected",
				lineNumber, cmd, paramsReceived, paramsExpected)
		}

		err = parserDescriptor.parser(params, sheet)
		if err != nil {
			return nil, fmt.Errorf("line %d: failed to parse %s command. %s", lineNumber, cmd, err.Error())
		}
	}

	dLen := len(durations)

	for fi, f := range sheet.Files {
		if dLen > fi {
			f.Duration = durations[fi]
		}
		for ti, t := range f.Tracks {
			t.StartPosition = t.StartTime().Seconds()
			var nextStart float64
			if len(f.Tracks) > ti+1 {
				nt := f.Tracks[ti+1]
				if nextStart = nt.Pregap.Seconds(); nextStart == 0 {
					nextStart = nt.StartTime().Seconds()
				}
			} else {
				nextStart = f.Duration
			}
			t.EndPosition = nextStart
		}
	}

	return sheet, nil
}

// parseCatalog parsers CATALOG command.
func parseCatalog(params []string, sheet *Sheet) error {
	num := params[0]
	matched, _ := regexp.MatchString("^[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]$", num)
	if !matched {
		return fmt.Errorf("%s is not valid catalog number", params)
	}
	sheet.Catalog = num
	return nil
}

// parseCdTextFile parsers CDTEXTFILE command.
func parseCdTextFile(params []string, sheet *Sheet) error {
	sheet.CdTextFile = params[0]
	return nil
}

// parseFile parsers FILE command.
// params[0] -- fileName
// params[1] -- fileType
func parseFile(params []string, sheet *Sheet) error {
	// Type parser function.
	parseFileType := func(t string) (fileType FileType, err error) {
		var types = map[string]FileType{
			"BINARY":   FileTypeBinary,
			"MOTOROLA": FileTypeMotorola,
			"AIFF":     FileTypeAiff,
			"WAVE":     FileTypeWave,
			"MP3":      FileTypeMp3,
		}

		fileType, ok := types[t]
		if !ok {
			err = fmt.Errorf("unknown file type: %s", t)
		}

		return
	}

	fileType, err := parseFileType(params[1])
	if err != nil {
		return err
	}

	file := *new(File)
	file.Name = params[0]
	file.Type = fileType

	sheet.Files = append(sheet.Files, &file)

	return nil
}

// parseFlags parsers FLAGS command.
func parseFlags(params []string, sheet *Sheet) error {
	flagParser := func(flag string) (trackFlag TrackFlag, err error) {
		var flags = map[string]TrackFlag{
			"DCP":  TrackFlagDcp,
			"4CH":  TrackFlag4ch,
			"PRE":  TrackFlagPre,
			"SCMS": TrackFlagScms,
		}

		trackFlag, ok := flags[flag]
		if !ok {
			err = fmt.Errorf("unknown track flag: %s", flag)
		}

		return
	}

	track := getCurrentTrack(sheet)
	if track == nil {
		return errors.New("TRACK command should appears before FLAGS command")
	}

	for _, flagStr := range params {
		flag, err := flagParser(flagStr)
		if err != nil {
			return err
		}
		track.Flags = append(track.Flags, flag)
	}

	return nil
}

// parseIndex parsers INDEX command.
func parseIndex(params []string, sheet *Sheet) error {
	min, sec, frames, err := parseTime(params[1])
	if err != nil {
		return errors.Wrap(err, "failed to parse index start time")
	}

	number, err := strconv.Atoi(params[0])
	if err != nil {
		return errors.Wrap(err, "failed to parse index number")
	}

	// All index numbers must be between 0 and 99 inclusive.
	if number < 0 || number > 99 {
		return errors.New("index number should be in 0..99 interval")
	}

	track := getCurrentTrack(sheet)
	if track == nil {
		return errors.New("TRACK command should appears before INDEX command")
	}

	// The first index of a file must start at 00:00:00.
	// if getFileLastIndex(getCurrentFile(sheet)) == nil {
	// 	if min+sec+frames != 0 {
	// 		return errors.New("first track index must start at 00:00:00")
	// 	}
	// }

	// This is the first track index?
	if len(track.Indexes) == 0 {
		// The first index must be 0 or 1.
		if number >= 2 {
			return errors.New("first track index should has 0 or 1 index number")
		}
	} else {
		// All other indexes being sequential to the first one.
		numberExpected := track.Indexes[len(track.Indexes)-1].Number + 1
		if numberExpected != number {
			return fmt.Errorf("expected %d index number but %d recieved", numberExpected, number)
		}
	}

	index := Index{Number: number, Time: Time{min, sec, frames}}
	track.Indexes = append(track.Indexes, index)

	return nil
}

// parseIsrc parsers ISRC command.
func parseIsrc(params []string, sheet *Sheet) error {
	isrc := params[0]

	track := getCurrentTrack(sheet)
	if track == nil {
		return errors.New("TRACK command should appears before ISRC command")
	}

	if len(track.Indexes) != 0 {
		return errors.New("ISRC command must be specified before INDEX command")
	}

	re := "^[0-9a-zA-z][0-9a-zA-z][0-9a-zA-z][0-9a-zA-z][0-9a-zA-z]" +
		"[0-9][0-9][0-9][0-9][0-9][0-9][0-9]$"
	matched, _ := regexp.MatchString(re, isrc)
	if !matched {
		return fmt.Errorf("%s is not valid ISRC number", isrc)
	}

	track.Isrc = isrc

	return nil
}

// parsePerformer parsers PERFORMER command.
func parsePerformer(params []string, sheet *Sheet) error {
	// Limit this field length up to 80 characters.
	performer := stringTruncate(params[0], 80)
	track := getCurrentTrack(sheet)

	if track == nil {
		// Performer command for the CD disk.
		sheet.Performer = performer
	} else {
		// Performer command for track.
		track.Performer = performer
	}

	return nil
}

// parsePostgap parsers POSTGAP command.
func parsePostgap(params []string, sheet *Sheet) error {
	track := getCurrentTrack(sheet)
	if track == nil {
		return errors.New("POSTGAP command must appear after a TRACK command")
	}

	min, sec, frames, err := parseTime(params[0])
	if err != nil {
		return errors.Wrap(err, "failed to parse postgap time")
	}

	track.Postgap = Time{min, sec, frames}

	return nil
}

// parsePregap parsers PREGAP command.
func parsePregap(params []string, sheet *Sheet) error {
	track := getCurrentTrack(sheet)
	if track == nil {
		return errors.New("PREGAP command must appear after a TRACK command")
	}

	if len(track.Indexes) != 0 {
		return errors.New("PREGAP command must appear before any INDEX command")
	}

	min, sec, frames, err := parseTime(params[0])
	if err != nil {
		return errors.Wrap(err, "failed to parse pregap time")
	}

	track.Pregap = Time{min, sec, frames}

	return nil
}

// parseRem parsers REM command.
func parseRem(params []string, sheet *Sheet) error {
	sheet.Comments = append(sheet.Comments, strings.Join(params, " "))

	return nil
}

// parseSongWriter parsers SONGWRITER command.
func parseSongWriter(params []string, sheet *Sheet) error {
	// Limit this field length up to 80 characters.
	songwriter := stringTruncate(params[0], 80)
	track := getCurrentTrack(sheet)

	if track == nil {
		sheet.Songwriter = songwriter
	} else {
		track.Songwriter = songwriter
	}

	return nil
}

// parseTitle parsers TITLE command.
func parseTitle(params []string, sheet *Sheet) error {
	// Limit this field length up to 80 characters.
	title := stringTruncate(params[0], 80)
	track := getCurrentTrack(sheet)

	if track == nil {
		// Title for the CD disk.
		sheet.Title = title
	} else {
		// Title command for track.
		track.Title = title
	}

	return nil
}

// parseTrack parses TRACK command.
func parseTrack(params []string, sheet *Sheet) error {
	fLen := len(sheet.Files)
	// TRACK command should be after FILE command.
	if fLen == 0 {
		return errors.New("unexpected TRACK command, FILE command expected first")
	}

	numberStr := params[0]
	dataTypeStr := params[1]

	// Type parser function.
	parseDataType := func(t string) (dataType TrackDataType, err error) {
		var (
			ok    bool
			types = map[string]TrackDataType{
				"AUDIO":      DataTypeAudio,
				"CDG":        DataTypeCdg,
				"MODE1/2048": DataTypeMode1_2048,
				"MODE1/2352": DataTypeMode1_2352,
				"MODE2/2336": DataTypeMode2_2336,
				"MODE2/2352": DataTypeMode2_2352,
				"CDI/2336":   DataTypeCdi_2336,
				"CDI/2352":   DataTypeCdi_2352,
			}
		)

		if dataType, ok = types[t]; !ok {
			err = fmt.Errorf("unknown track datatype: %s", t)
		}
		return
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return errors.Wrap(err, "failed to parse track number parameter")
	}
	if number < 1 {
		return errors.New("failed to parse track number parameter. value should be in 1..99 range")
	}

	dataType, err := parseDataType(dataTypeStr)
	if err != nil {
		return err
	}

	track := new(Track)
	track.Number = number
	track.DataType = dataType

	file := sheet.Files[fLen-1]
	tLen := len(file.Tracks)

	// But all track numbers after the first must be sequential.
	if tLen > 0 {
		if file.Tracks[tLen-1].Number != number-1 {
			return fmt.Errorf("expected track number %d, but %d recieved", number-1, number)
		}
	}

	file.Tracks = append(file.Tracks, track)

	return nil
}

// getCurrentFile returns file object started with the last FILE command.
// Returns nil if there is no any File objects.
func getCurrentFile(sheet *Sheet) (f *File) {
	if tLen := len(sheet.Files); tLen > 0 {
		f = sheet.Files[tLen-1]
	}
	return
}

// getCurrentTrack returns current track object, which was started with last TRACK command.
// Returns nil if there is no any Track object available.
func getCurrentTrack(sheet *Sheet) (t *Track) {
	file := getCurrentFile(sheet)
	if file != nil {
		if tLen := len(file.Tracks); tLen > 0 {
			t = file.Tracks[tLen-1]
		}
	}
	return
}

// getFileLastIndex returns last index for the given file.
// Returns nil if file has no any indexes.
func getFileLastIndex(file *File) *Index {
	for _, t := range file.Tracks {
		if l := len(t.Indexes); l > 0 {
			return &(t.Indexes[l-1])
		}
	}
	return nil
}
