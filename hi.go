package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

import (
	"github.com/as/hue"
)

type runtimeArgs struct {
	fg, bg, regexp, pattern *string
	files                   []*string
}

// badColor lets the user know that the color they selected is unsupported. It iterates through all
// of the supported colors in the hue package and prints them to the screen.
func badColor(color string) {
	fmt.Printf("Error: \"%s\" is not a supported color.\nThe following colors are supported:\n%s\n", color, func() (s string) {
		for i := hue.First; i < hue.Last; i++ {
			x := hue.New(i, hue.Default)
			if i == hue.Black {
				x.SetBg(hue.White)
			}
			s += string(hue.Encode(x, hue.HueToString[i])) + " "
		}
		return s
	}())
}

func showHelp() {
	fmt.Printf("Usage: hi [OPTIONS]... PATTERN [FILE]...\n\n")
	fmt.Printf("OPTIONS\n")
	flag.PrintDefaults()
	fmt.Printf("\nPATTERN\n")
	fmt.Printf("A POSIX regular expression\n")
	fmt.Printf("\nFILE\n")
	fmt.Printf("One or more files\n\n")
	fmt.Printf("EXAMPLES\n")
	fmt.Printf("ifconfig | hi 'inet .*'\n")
	fmt.Printf("hi --fg blue --bg white defaults < /etc/fstab\n")
	fmt.Printf("hi '[eE]' ~/books/thegreatgatsby.txt\n")
	fmt.Printf("\nRETURN VALUE\n")
	fmt.Printf("Returns 0 if no fatal errors have occured and at least one pattern is matched\n")
}

// validArgs returns true if the required arguments to run the program have been provided
// on the command line.
func validArgs(args *runtimeArgs) bool {
	flag.Usage = showHelp

	args.fg = flag.String("fg", "black", "foreground color")
	args.bg = flag.String("bg", "green", "background color")

	flag.Parse()

	if args.fg == nil || hue.StringToHue[*args.fg] == 0 {
		badColor(*args.fg) // prints an error message w/ a list of supported colors
		return false
	}

	if args.fg == nil || hue.StringToHue[*args.bg] == 0 {
		badColor(*args.bg)
		return false
	}

	// Get the remaining flags, which should
	// consist of a pattern, and optionally, one or more file names.
	rem := flag.Args()

	switch {
	case len(rem) == 0:
		fmt.Println("Error: No pattern specified.")
		showHelp()
		return false
	case len(rem) == 1:
		args.pattern = &rem[0]
	case len(rem) >= 2:
		args.pattern = &rem[0]

		for i := 1; i < len(rem); i++ {
			args.files = append(args.files, &rem[i])
		}
	}

	return true
}

func pipedWrite(in *bufio.Reader, out *hue.RegexpWriter) int {
	var (
		buf        []byte
		err        error
		r, w, tmpw int
	)

	for {
		if buf, err = in.ReadBytes('\n'); err != nil {
			break
		}
		r += len(buf)

		if tmpw, err = out.Write(buf); err != nil {
			break
		}
		w += tmpw
	}

	return w - r
}

func main() {
	var args runtimeArgs

	// Difference between bytes written and bytes read. If delta != 0,
	// that means a match for the PATTERN was found, and the program will
	// return 0.
	var delta int

	if !validArgs(&args) {
		os.Exit(1)
	}

	fg := hue.StringToHue[*args.fg]
	bg := hue.StringToHue[*args.bg]

	// Configure colors and output stuff
	h := hue.New(fg, bg)
	out := hue.NewRegexpWriter(os.Stdout)
	out.AddRuleStringPOSIX(h, *args.pattern)

	// Like grep, if no files are specified, we read from stdin
	if len(args.files) == 0 {
		in := bufio.NewReader(os.Stdin)
		delta = pipedWrite(in, out)
	} else {
		// Otherwise we open the files one-by-one
		for _, v := range args.files {
			fd, err := os.Open(*v)
			if err != nil {
				fmt.Println(err)
				continue
			}

			in := bufio.NewReader(fd)
			delta += pipedWrite(in, out)
		}
	}

	if delta == 0 {
		// No matches found for pattern
		os.Exit(3)
	}

	os.Exit(0)
}
