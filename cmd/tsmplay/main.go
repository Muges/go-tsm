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
	"fmt"
	"github.com/Muges/tsm"
	"github.com/Muges/tsm/streamer"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"time"
)

var (
	app = kingpin.New("tsmplay", "Change the speed of a WAV audio file.")

	speed          = app.Flag("speed", "Change the speed by N percents (100 by default).").Short('s').PlaceHolder("N").Int()
	frameSize      = app.Flag("frame_size", "Set the frame size to N.").Short('f').PlaceHolder("N").Int()
	synthesisHop   = app.Flag("synthesis_hop", "Set the synthesis hop to N.").PlaceHolder("N").Int()
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

	// Set default values
	if *speed == 0 {
		*speed = 100
	}
	if *frameSize == 0 {
		*frameSize = format.SampleRate.N(10 * time.Millisecond)
	}
	if *synthesisHop == 0 {
		*synthesisHop = *frameSize / 4
	}

	analysisHop := (*synthesisHop * *speed) / 100

	t := tsm.New(2, analysisHop, *synthesisHop, *frameSize, *frameSize)
	stretchedStream := streamer.New(&t, stream)

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
