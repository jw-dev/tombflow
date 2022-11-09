package script

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	LEnglish Language = iota
	LFrench
	LGerman
	LAmerican
	LJapanese
	LItalian
	LSpanish
)

const (
	OpPicture Opcode = iota
	OpListStart
	OpListEnd
	OpFmv
	OpLevel
	OpCine
	OpComplete
	OpDemo
	OpJumpToSequence
	OpEnd
	OpTrack
	OpSunset
	OpLoadPic
	OpDeadlyWater
	OpRemoveWeapons
	OpGameComplete
	OpCutAngle
	OpNoFloor
	OpStartInv
	OpStartAnim
	OpSecrets
	OpKillToComplete
	OpRemoveAmmo
)

const (
	ELevel Event = 1000 + iota
	ESavedGame
	ECutscene
	EFmv
	EDemo
	EExittoTitle
	EExitGame
)

var tLanguages = [...]string{
	"English",
	"French",
	"German",
	"American",
	"Japanese",
	"Italian",
	"Spanish",
}

// tOpcodes is a collection of English-language translations for gameflow opcodes.
var tOpcodes = [...]string{
	"Picture",
	"List Start",
	"List End",
	"Display FMV",
	"Play Level",
	"Display Cutscene",
	"End Level",
	"Play Demo",
	"Jump to Sequence",
	"End Sequence",
	"Play Soundtrack",
	"Sunset",
	"Load Pic",
	"Deadly Water",
	"Remove Weapons",
	"Game Complete",
	"Set Cutscene Angle",
	"No Floor",
	"Start Inv",
	"Start Animation",
	"Secrets",
	"Kill to Complete",
	"Remove Ammo",
}

// tEvents is a collection of English-language translations for script events.
var tEvents = [...]string{
	"Load Level",
	"Load Saved Game",
	"Load Cutscene",
	"Load FMV",
	"Load Random Demo",
	"Exit to Title",
	"Exit Game",
}

// hasOpcodeArgument is a list of opcodes that have some argument.
// This is needed because when parsing the gameflow, the following uint16 is the argument only for these opcodes.
var hasOpcodeArgument = []Opcode{
	OpPicture,
	OpFmv,
	OpLevel,
	OpCine,
	OpDemo,
	OpJumpToSequence,
	OpTrack,
	OpLoadPic,
	OpCutAngle,
	OpNoFloor,
	OpStartInv,
	OpStartAnim,
	OpSecrets,
}

// hasEventArgument is a list of Events that have an argument.
var hasEventArgument = []Event{
	ELevel,
	ESavedGame,
	ECutscene,
	EFmv,
}

type Language uint8

func (l Language) String() string {
	if int(l) < len(tLanguages) {
		return tLanguages[l]
	}
	return "Unknown"
}

type Opcode uint16

func (o Opcode) String() string {
	if int(o) < len(tOpcodes) {
		return tOpcodes[o]
	}
	return "Unknown"
}

func (o Opcode) hasArg() bool {
	for i := range hasOpcodeArgument {
		if o == hasOpcodeArgument[i] {
			return true
		}
	}
	return false
}

type Event uint16

func (e Event) String() string {
	if int(e) < len(tEvents) {
		return tEvents[e]
	}
	return "Unknown"
}

func (e Event) HasArg() bool {
	for i := range hasEventArgument {
		if e == hasEventArgument[i] {
			return true
		}
	}
	return false
}

type Ev struct {
	Typ Event
	Arg uint8
}

type Op struct {
	Typ Opcode
	Arg uint16
}

type Sequence []Op

type Level struct {
	Name    string
	Path    string
	Chapter string
	Flow    Sequence
	IsDemo  bool
	Puzzles [4]string
	Pickups [2]string
	Keys    [4]string
}

type Script struct {
	Version      uint32
	Description  string
	Lang         Language
	Levels       []Level
	Titles       []string
	Fmvs         []string
	Cutscenes    []string
	GameStrings  []string
	ExtraStrings []string
}

func Read(r io.Reader) *Script {
	head := readHeader(r)
	levelNames := readStringArray(r, head.NumLevels, head.XorKey)
	chapterPaths := readStringArray(r, head.NumChapterScreens, head.XorKey)
	titlePaths := readStringArray(r, head.NumTitles, head.XorKey)
	fmvPaths := readStringArray(r, head.NumFmvs, head.XorKey)
	levelPaths := readStringArray(r, head.NumLevels, head.XorKey)
	cutscenePaths := readStringArray(r, head.NumCutscenes, head.XorKey)
	gameFlow := readSequenceArray(r, head.NumLevels+1)
	demoLevels := readDemoLevels(r, head.NumDemoLevels)
	gameStrings := readGameStrings(r, head.XorKey)
	extraStrings := readStringArray(r, 41, head.XorKey)
	levels := joinLevels(levelNames, levelPaths, chapterPaths, gameFlow, demoLevels)

	for i := 0; i < 4; i++ {
		puzzles := readStringArray(r, head.NumLevels, head.XorKey)
		for j := 0; j < int(head.NumLevels); j++ {
			levels[j].Puzzles[i] = puzzles[j]
		}
	}

	for i := 0; i < 2; i++ {
		puzzles := readStringArray(r, head.NumLevels, head.XorKey)
		for j := 0; j < int(head.NumLevels); j++ {
			levels[j].Pickups[i] = puzzles[j]
		}
	}

	for i := 0; i < 4; i++ {
		puzzles := readStringArray(r, head.NumLevels, head.XorKey)
		for j := 0; j < int(head.NumLevels); j++ {
			levels[j].Keys[i] = puzzles[j]
		}
	}

	return &Script{
		Version:      head.Version,
		Description:  string(head.Description[:]),
		Levels:       levels,
		Titles:       titlePaths,
		Fmvs:         fmvPaths,
		Cutscenes:    cutscenePaths,
		GameStrings:  gameStrings,
		ExtraStrings: extraStrings,
	}
}

// FormatOpcode properly formats an Opcode, substituting args for their appropriate values.
// For example, "LoadLevel 0" would be formatted as "Load Level: JUNGLE.TR2 (Jungle)" instead (in TRIII).
func (s Script) FormatOp(o Op) string {
	if !o.Typ.hasArg() {
		return fmt.Sprintf("%v", o.Typ)
	}
	switch o.Typ {
	case OpLoadPic:
		return fmt.Sprintf("%v '%v'", o.Typ, s.Levels[o.Arg].Chapter)
	case OpFmv:
		return fmt.Sprintf("%v '%v'", o.Typ, s.Fmvs[o.Arg])
	case OpLevel:
		level := s.Levels[o.Arg]
		return fmt.Sprintf("%v '%v' (%v)", o.Typ, level.Path, level.Name)
	case OpCine:
		return fmt.Sprintf("%v '%v'", o.Typ, s.Cutscenes[o.Arg])
	default:
		return fmt.Sprintf("%v %v", o.Typ, o.Arg)
	}
}

type header struct {
	Version           uint32
	Description       [256]byte
	GameflowSize      uint16
	FirstOption       int32
	TitleReplace      int32
	OnDeathDemoMode   int32
	OnDeathInGame     int32
	DemoTime          uint32
	OnDemoInterrupt   int32
	OnDemoEnd         int32
	_                 [36]byte
	NumLevels         uint16
	NumChapterScreens uint16
	NumTitles         uint16
	NumFmvs           uint16
	NumCutscenes      uint16
	NumDemoLevels     uint16
	TitleSoundId      uint16
	SingleLevel       uint16
	_                 [32]byte
	Flags             uint16
	_                 [6]byte
	XorKey            byte
	LanguageId        byte
	SecretSoundId     uint16
	_                 [4]byte
}

func readHeader(r io.Reader) *header {
	h := header{}
	binary.Read(r, binary.LittleEndian, &h)
	return &h
}

func readOffsetsAndSize(r io.Reader, count uint16) (off []uint16, sz uint16) {
	off = make([]uint16, count)
	binary.Read(r, binary.LittleEndian, &off)
	binary.Read(r, binary.LittleEndian, &sz)
	return
}

func readStringArray(r io.Reader, count uint16, xor byte) []string {
	strs := make([]string, 0)
	off, sz := readOffsetsAndSize(r, count)
	data := make([]byte, sz)
	binary.Read(r, binary.LittleEndian, &data)
	if xor > 0 {
		for i := range data {
			data[i] ^= xor
		}
	}
	for i, offset := range off {
		if i == len(off)-1 {
			strs = append(strs, string(data[offset:]))
			break
		}
		to := off[i+1]
		strs = append(strs, string(data[offset:to]))
	}
	return strs
}

func readSequenceArray(r io.Reader, count uint16) []Sequence {
	seqs := []Sequence{}
	chunks := make([][]uint16, 0)

	off, size := readOffsetsAndSize(r, count)
	data := make([]uint16, size/2)
	binary.Read(r, binary.LittleEndian, &data)

	for i, offset := range off {
		if i == len(off)-1 {
			chunks = append(chunks, data[offset/2:])
			break
		}
		to := off[i+1]
		chunks = append(chunks, data[offset/2:to/2])
	}

	for _, chunk := range chunks {
		seq := Sequence{}

		for i := 0; i < len(chunk); i++ {
			typ := Opcode(chunk[i])
			arg := uint16(0)
			if typ.hasArg() {
				i = i + 1
				arg = chunk[i]
			}
			seq = append(seq, Op{Typ: typ, Arg: arg})
		}

		seqs = append(seqs, seq)
	}

	return seqs
}

func readDemoLevels(r io.Reader, count uint16) []uint16 {
	levels := make([]uint16, count)
	binary.Read(r, binary.LittleEndian, &levels)
	return levels
}

func readGameStrings(r io.Reader, xor byte) []string {
	count := uint16(0)
	binary.Read(r, binary.LittleEndian, &count)
	return readStringArray(r, count, xor)
}

func joinLevels(names []string, paths []string, chaps []string, flow []Sequence, demos []uint16) []Level {
	l := make([]Level, len(names))

	for i, name := range names {
		l[i].Name = name

		if i < len(paths) {
			l[i].Path = paths[i]
		}
		if i < len(chaps) {
			l[i].Chapter = chaps[i]
		}
		if i+1 < len(flow) {
			l[i].Flow = flow[i+1]
		}
	}

	for _, demo := range demos {
		if int(demo) < len(l) {
			l[demo].IsDemo = true
		}
	}

	return l
}
