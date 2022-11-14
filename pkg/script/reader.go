package script

import (
	"encoding/binary"
	"io"
)

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

func readHeader(r io.Reader) *header {
	h := header{}
	binary.Read(r, binary.LittleEndian, &h)
	return &h
}

func readMultiByteArray(r io.Reader, count uint16) *multiByteArray {
	offsets := make([]uint16, count)
	binary.Read(r, binary.LittleEndian, &offsets)

	size := uint16(0)
	binary.Read(r, binary.LittleEndian, &size)

	data := make([]uint8, size)
	binary.Read(r, binary.LittleEndian, &data)

	return newMultiByteArray(offsets, data)
}

func readStringArray(r io.Reader, count uint16, xor byte) []string {
	m := readMultiByteArray(r, count)
	return m.Strings(xor)
}

func readSequenceArray(r io.Reader, count uint16) []Sequence {
	seqs := []Sequence{}

	m := readMultiByteArray(r, count)
	chunks := m.U16()

	for _, chunk := range chunks {
		seq := Sequence{}

		for i := 0; i < len(chunk); i++ {
			typ := Opcode(chunk[i])
			arg := uint16(0)
			if typ.hasArg() {
				i = i + 1
				arg = chunk[i]
			}
			seq = append(seq, Command{Op: typ, Arg: arg})
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
