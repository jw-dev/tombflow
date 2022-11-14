package script

import (
	"encoding/binary"
	"fmt"
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
	OpSavedGame
	OpExitToTitle
	OpExitGame
	OpDisable = -1
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

var commandOpcodeMap = map[uint16]Opcode{
	0: OpLevel,
	1: OpSavedGame,
	2: OpCine,
	3: OpFmv,
	4: OpDemo,
	5: OpExitToTitle,
	6: OpExitGame,
}

var opcodeHasArgument = []Opcode{
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

type Language uint8

func (l Language) String() string {
	if int(l) < len(tLanguages) {
		return tLanguages[l]
	}
	return "Unknown"
}

type Opcode int32

func (o Opcode) hasArg() bool {
	for i := 0; i < len(opcodeHasArgument); i++ {
		if opcodeHasArgument[i] == o {
			return true
		}
	}
	return false
}

func (o Opcode) String() string {
	if int(o) < len(tOpcodes) {
		return tOpcodes[o]
	}
	return "Unknown"
}

type Command struct {
	Op  Opcode
	Arg uint16
}

func (c Command) String() string {
	if c.Op.hasArg() {
		return fmt.Sprintf("%v %v", c.Op, c.Arg)
	}
	return c.Op.String()
}

type Sequence []Command

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

// FormatCommand formats a comand, replacing any arguments with the relevent item. For example, (LoadLevel 0) would return "Load Level JUNGLE.PSX" (in TRIII)
func (s Script) FormatCommand(c Command) string {
	if !c.Op.hasArg() {
		return c.Op.String()
	}
	switch c.Op {
	case OpLoadPic:
		return fmt.Sprintf("%v '%v'", c.Op, s.Levels[c.Arg].Chapter)
	case OpFmv:
		return fmt.Sprintf("%v '%v'", c.Op, s.Fmvs[c.Arg])
	case OpLevel:
		level := s.Levels[c.Arg]
		return fmt.Sprintf("%v '%v' (%v)", c.Op, level.Path, level.Name)
	case OpCine:
		return fmt.Sprintf("%v '%v'", c.Op, s.Cutscenes[c.Arg])
	default:
		return fmt.Sprintf("%v %v", c.Op, c.Arg)
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

type multiByteArray struct {
	offsets []uint16
	data    []uint8
}

func newMultiByteArray(offsets []uint16, data []uint8) *multiByteArray {
	return &multiByteArray{
		offsets,
		data,
	}
}

func (m multiByteArray) Strings(xor byte) []string {
	strs := []string{}

	for i, offset := range m.offsets {
		bytes := []uint8{}
		switch i {
		case len(m.offsets) - 1:
			bytes = m.data[offset:]
		default:
			to := m.offsets[i+1]
			bytes = m.data[offset:to]
		}
		if xor > 0 {
			for i := range bytes {
				bytes[i] ^= xor
			}
		}
		strs = append(strs, string(bytes))
	}

	return strs
}

func (m multiByteArray) U16() [][]uint16 {
	u16 := make([]uint16, len(m.data)/2)

	for i := 0; i < len(u16); i++ {
		u16[i] = binary.LittleEndian.Uint16(m.data[i*2:])
	}

	chunks := [][]uint16{}

	for i, offset := range m.offsets {
		if i == len(m.offsets)-1 {
			chunks = append(chunks, u16[offset/2:])
			break
		}
		to := m.offsets[i+1]
		chunks = append(chunks, u16[offset/2:to/2])
	}

	return chunks
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
