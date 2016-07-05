package main

import (
	"flag"
	"fmt"
	"image"
	png "image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/nfnt/resize"
)

func main() {
	commands := map[string]command{
		"create": createCmd(),
	}

	fs := flag.NewFlagSet("appiconizer", flag.ExitOnError)
	cpus := fs.Int("cpus", runtime.NumCPU(), "Number of CPUs to use")
	profile := fs.String("profile", "", "Enable profiling of [cpu, heap]")
	version := fs.Bool("version", false, "Print version and exit")

	fs.Usage = func() {
		fmt.Println("Usage: appiconizer  <command> [command flags]")
		for name, cmd := range commands {
			fmt.Printf("\n%s command:\n", name)
			cmd.fs.PrintDefaults()
		}
		fmt.Println(examples)
	}

	fs.Parse(os.Args[1:])

	if *version {
		fmt.Println(Version)
		return
	}

	runtime.GOMAXPROCS(*cpus)

	for _, prof := range strings.Split(*profile, ",") {
		if prof = strings.TrimSpace(prof); prof == "" {
			continue
		}

		f, err := os.Create(prof + ".pprof")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		switch {
		case strings.HasPrefix(prof, "cpu"):
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		case strings.HasPrefix(prof, "heap"):
			defer pprof.Lookup("heap").WriteTo(f, 0)
		}
	}

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		os.Exit(1)
	}

	if cmd, ok := commands[args[0]]; !ok {
		log.Fatalf("Unknown command: %s", args[0])
	} else if err := cmd.fn(args[1:]); err != nil {
		log.Fatal(err)
	}
}

func createCmd() command {
	fs := flag.NewFlagSet("appiconizer create", flag.ExitOnError)
	opts := &options{}
	fs.StringVar(&opts.file, "source", "", "Source file")
	fs.StringVar(&opts.device, "device", "all", "ios/android/all")
	return command{fs, func(args []string) error {
		fs.Parse(args)
		return create(opts)
	}}
}

func create(opts *options) (err error) {
	fmt.Println(opts)

	file, err := os.Open(opts.file)
	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Dir(opts.file)
	fmt.Println(filePath)

	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	var sizes []uint

	if opts.device == "ios" {
		sizes = []uint{29, 58, 87, 80, 120, 120, 180, 40, 76, 152, 167}
	} else if opts.device == "android" {
		sizes = []uint{48, 72, 96, 144, 192}
	} else {
		sizes = []uint{29, 58, 87, 80, 120, 120, 180, 40, 76, 152, 167, 48, 72, 96, 144, 192}
	}

	for _, val := range sizes {
		resizeImage(uint(val), img)
	}

	fmt.Println("***Done***")
	return nil
}

func resizeImage(width uint, img image.Image) {

	name := fmt.Sprintf("icon_%d.png", width)
	m := resize.Resize(width, 0, img, resize.Lanczos3)
	out, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}

	defer out.Close()

	// write new image to file
	err = png.Encode(out, m)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("created", name)
}

// Version is set at compile time.
var Version = "???"

const examples = `
examples:
  appiconizer create -source icon.png
`

type options struct {
	file   string
	device string
}

type command struct {
	fs *flag.FlagSet
	fn func(args []string) error
}
