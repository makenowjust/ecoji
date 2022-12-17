//go:build ignore
// +build ignore

// This program generates emojis.go. It can be invoked by running
// go run gen.go
package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
)

func main() {
	user := os.Getenv("USER")
	if user == "" {
		user = "robots"
	}
	v1Lines := getLines("../emojisV1.txt")
	v2Lines := getLines("../emojisV2.txt")

	revMap := make(map[string]RuneInfo)

	for i, r := range v1Lines {
		revMap[r] = RuneInfo{
			Ordinal: i,
			Version: 1,
		}
	}

	for i, r := range v2Lines {
		rInfo, exists := revMap[r]

		if exists {
			if rInfo.Ordinal != i {
				panic("Ordinal mismatch " + string(i) + " " + string(r))
			}

			rInfo.Version = 3
			revMap[r] = rInfo
		} else {
			revMap[r] = RuneInfo{
				Ordinal: i,
				Version: 2,
			}
		}

	}

	doc := document{
		User:      user,
		Timestamp: time.Now().Format(time.RFC3339),
		EmojisV1:  v1Lines,
		EmojisV2:  v2Lines,
		RevMap:    revMap,
	}
	ef, err := os.Create("mapping.go")
	handle(err)
	defer ef.Close()
	handle(mappingTemplate.Execute(ef, doc))
}

func getLines(fileName string) []string {
	buf, err := os.ReadFile(fileName)
	handle(err)
	return strings.Split(strings.ToLower(strings.TrimSpace(string(buf))), "\n")
}

func handle(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type document struct {
	User      string
	Timestamp string
	EmojisV1  []string
	EmojisV2  []string
	RevMap    map[string]RuneInfo
}

type RuneInfo struct {
	Ordinal int
	Version int
}

var mappingTemplate = template.Must(template.New("").Parse(`// Code generated; DO NOT EDIT.
// This file was generated by {{ .User }} at
// {{ .Timestamp }}
package ecoji


// padding to use when less than 5 bytes are present to encode
const padding rune = 0x2615

// padding to use for the last emoji when only 4 input bytes are preset
var paddingLastV1 = [4]rune{0x269C,0x1F3CD,0x1F4D1,0x1F64B}
var paddingLastV2 = [4]rune{0x1F977,0x1F6FC,0x1F4D1,0x1F64B}

// The emojis used for ecoji version 1
var emojisV1 = [1024]rune{
{{- range $i, $emoji := .EmojisV1 }}
	0x{{$emoji}},
{{- end }}
}

// The emojis used for ecoji version 2
var emojisV2 = [1024]rune{
{{- range $i, $emoji := .EmojisV2 }}
	0x{{$emoji}},
{{- end }}
}

type ecojiver int

const (
    // signifies that an emoji is only used by ecoji version 1
	ev1    ecojiver = 1
    // signifies that an emoji is only used by ecoji version 2
	ev2    ecojiver = 2
    // signifies that an emoji is used by ecoji version 1 and 2
	evAll  ecojiver = 3
)

type paddingType int

const (
    // This indicates an emoji is not used for padding
	padNone = -1
    // This indicates the primary emoji padding chararacter
    padFill = 1
    // This indicates one of the special padding characters used for encoding 4 byte data
    padLast = 2
)

// This struct contains information about an emoji that ecoji uses for decoding
type emojiInfo struct {
    // This is the 10-bit code that an emoji maps to for decoding
	ordinal int
    // This indicates what version of ecoji use the emoji
	version ecojiver
    // This indicates if an emojis used for padding
    padding paddingType
}

// This map is used for ecoji decoding. It maps emojis to the information needed to decode them into 10-bit integers. 
// It also has information needed to decoding padding emojis and validate ecoji version 1 and version 2.
var revEmojis = map[rune]emojiInfo{
{{- range $r, $ri := .RevMap }}
	0x{{$r}}: { ordinal: {{$ri.Ordinal}}, version: {{$ri.Version}}, padding: padNone },
{{- end }}
    padding: { ordinal: 0, version: evAll, padding: padFill }, 
    paddingLastV1[0]: { ordinal: 0, version: ev1, padding: padLast }, 
    paddingLastV1[1]: { ordinal: 1<<8, version: ev1, padding: padLast }, 
    paddingLastV1[2]: { ordinal: 2<<8, version: evAll, padding: padLast }, 
    paddingLastV1[3]: { ordinal: 3<<8, version: evAll, padding: padLast },
    paddingLastV2[0]: { ordinal: 0, version: ev2, padding: padLast },
    paddingLastV2[1]: { ordinal: 1<<8, version: ev2, padding: padLast },
}

`))