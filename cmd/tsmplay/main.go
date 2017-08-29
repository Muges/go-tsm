// Copyright (c) 2017 Muges
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package main

import (
	"errors"
	"fmt"
	"github.com/Muges/tsm/ola"
	"github.com/Muges/tsm/streamer"
	"github.com/Muges/tsm/tsm"
	"github.com/Muges/tsm/wsola"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"time"
)

var (
	app = kingpin.New("tsmplay", "Change the speed of a WAV audio file.")

	speed          = app.Flag("speed", "Change the speed by N percents (100 by default).").Short('s').PlaceHolder("N").Default("-1").Float64()
	method         = app.Flag("method", "Change the TSM method (ola or wsola).").Short('m').PlaceHolder("METHOD").Default("wsola").Enum("ola", "wsola")
	frameLength    = app.Flag("frame_length", "Set the frame length to N.").Short('l').PlaceHolder("N").Default("-1").Int()
	synthesisHop   = app.Flag("synthesis_hop", "Set the synthesis hop to N.").PlaceHolder("N").Default("-1").Int()
	tolerance      = app.Flag("tolerance", "Set the tolerance for the WSOLA procedure to N.").Short('t').PlaceHolder("N").Default("-1").Int()
	outputFilename = app.Flag("output", "Save the stretched audio to FILENAME instead of playing it.").Short('o').PlaceHolder("FILENAME").String()

	inputFilename = app.Arg("filename", "A wav file.").Required().ExistingFile()
)

func main() {
	// Read command-line arguments
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Open and decode wav file
	inputFile, err := os.Open(*inputFilename)
	if err != nil {
		fmt.Printf("error: unable to open file \"%s\"\n", *inputFilename)
		fmt.Println(err)
		os.Exit(1)
	}
	defer inputFile.Close()

	stream, format, err := wav.Decode(inputFile)
	if err != nil {
		fmt.Printf("error: \"%s\" is not a valid wav file\n", *inputFilename)
		fmt.Println(err)
		os.Exit(1)
	}

	// Create TSM object
	var t *tsm.TSM
	switch *method {
	case "ola":
		t, err = ola.NewWithSpeed(2, *speed, *synthesisHop, *frameLength)
	case "wsola":
		t, err = wsola.NewWithSpeed(2, *speed, *synthesisHop, *frameLength, *tolerance)
	default:
		err = errors.New(fmt.Sprintf("Unknown TSM method \"%s\"", *method))
	}
	if err != nil {
		fmt.Println("error: unable to create the TSM object")
		fmt.Println(err)
		os.Exit(1)
	}
	stretchedStream := streamer.New(t, stream)

	if *outputFilename != "" {
		outputFile, err := os.Create(*outputFilename)
		if err != nil {
			fmt.Printf("error: unable to open file \"%s\"\n", *outputFilename)
			fmt.Println(err)
			os.Exit(1)
		}
		defer outputFile.Close()

		wav.Encode(outputFile, stretchedStream, format)
	} else {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		speaker.UnderrunCallback(func() { fmt.Println("underrun") })

		// Create a channel that will be closed at the end of playback
		done := make(chan struct{})

		speaker.Play(beep.Seq(&stretchedStream, beep.Callback(func() {
			close(done)
		})))

		// Wait for the channel to be closed before quitting
		<-done
	}
}
