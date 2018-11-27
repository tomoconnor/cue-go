package cue

const (
	framesPerSecond = 75

	// Intel binary file (least significant byte first)
	FileTypeBinary FileType = iota
	// Motorola binary file (most significant byte first)
	FileTypeMotorola
	// Audio AIFF file
	FileTypeAiff
	// Audio WAVE file
	FileTypeWave
	// Audio MP3 file
	FileTypeMp3

	// AUDIO – Audio/Music (2352)
	DataTypeAudio = iota
	// CDG – Karaoke CD+G (2448)
	DataTypeCdg
	// MODE1/2048 – CDROM Mode1 Data (cooked)
	DataTypeMode1_2048
	// MODE1/2352 – CDROM Mode1 Data (raw)
	DataTypeMode1_2352
	// MODE2/2336 – CDROM-XA Mode2 Data
	DataTypeMode2_2336
	// MODE2/2352 – CDROM-XA Mode2 Data
	DataTypeMode2_2352
	// CDI/2336 – CDI Mode2 Data
	DataTypeCdi_2336
	// CDI/2352 – CDI Mode2 Data
	DataTypeCdi_2352

	// Digital copy permitted.
	TrackFlagDcp = iota
	// Four channel audio.
	TrackFlag4ch
	// Pre-emphasis enabled (audio tracks only).
	TrackFlagPre
	// Serial copy management system (not supported by all recorders).
	TrackFlagScms
)

type (
	// Cue sheet file representation.
	Sheet struct {
		// Disc's media catalog number.
		Catalog string
		// Name of a perfomer for a CD-TEXT enhanced disc
		Performer string
		// Specify a title for a CD-TEXT enhanced disc.
		Title string
		// Specify songwriter for disc.
		Songwriter string
		// Comments in the CUE SHEET file.
		Comments []string
		// Name of the file that contains the encoded CD-TEXT information for the disc.
		CdTextFile string
		// Data/audio files descibed byt the cue-file.
		Files []*File
	}

	// Track datatype.
	TrackDataType int

	// Type of the audio file.
	FileType int

	// Additional decode information about track.
	TrackFlag int

	// Time point description type.
	Time struct {
		// Minutes.
		Min int
		// Minutes.
		Sec int
		// Frames.
		Frames int
	}

	// Track index type
	Index struct {
		// Index number.
		Number int
		// Index starting time.
		Time Time
	}

	Track struct {
		// Track number (1-99).
		Number int
		// Track datatype.
		DataType TrackDataType
		// Track title.
		Title string
		// Track preformer.
		Performer string
		// Songwriter.
		Songwriter string
		// Track decode flags.
		Flags []TrackFlag
		// Internetional Standaard Recording Code.
		Isrc string
		// Track indexes.
		Indexes []Index
		// Length of the track pregap.
		Pregap Time
		// Length of the track postgap.
		Postgap       Time
	}

	// Audio file representation structure.
	File struct {
		// Name (path) of the file.
		Name string
		// Type of the audio file.
		Type FileType
		// List of present tracks in the file.
		Tracks []*Track
	}
)

// Seconds returns length in seconds.
func (time Time) Seconds() float64 {
	return float64(time.Min*60) + float64(time.Sec) + float64(time.Frames)/framesPerSecond
}
