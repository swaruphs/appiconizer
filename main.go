package main

import (
	"archive/zip"
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
	"time"

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
	fs.StringVar(&opts.target, "target", "", "Target location")
	fs.BoolVar(&opts.zip, "zip", false, "zip files")
	return command{fs, func(args []string) error {
		fs.Parse(args)
		return create(opts)
	}}
}

func create(opts *options) (err error) {

	file, err := os.Open(opts.file)
	if err != nil {
		log.Fatal(err)
	}

	var fileName = file.Name()
	var extension = filepath.Ext(file.Name())
	var name = opts.file[0 : len(fileName)-len(extension)]
	opts.name = name

	path := opts.target
	if path == "" {
		path = filepath.Dir(opts.file)
	}

	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	var sizes []uint

	if opts.device == "ios" {
		sizes = []uint{29, 48, 55, 58, 87, 88, 80, 120, 180, 40, 76, 152, 167, 172, 196}
	} else if opts.device == "android" {
		sizes = []uint{48, 72, 96, 144, 192}
	} else {
		sizes = []uint{29, 48, 55, 58, 87, 88, 80, 120, 180, 40, 76, 152, 167, 48, 72, 96, 144, 192, 172, 196}
	}

	if opts.zip == true {
		zipFile(opts, sizes, &img, path)
	} else {
		dirPath := filepath.Join(path, getFolderName())
		os.Mkdir(dirPath, 0777)
		for _, val := range sizes {
			resizeImage(uint(val), img, dirPath)
		}
	}
	fmt.Println("***Done***")
	return nil
}

// Func to create zip file with list of all icons
func zipFile(opts *options, sizes []uint, img *image.Image, path string) {

	fileName := fmt.Sprintf("%s.%s", getFolderName(), "zip")
	zipPath := filepath.Join(path, fileName)
	writer, err := os.Create(zipPath)
	defer writer.Close()
	if err != nil {
		log.Fatal(err)
	}

	w := zip.NewWriter(writer)

	for _, width := range sizes {
		name := fmt.Sprintf("icon_%d.png", width)
		m := resize.Resize(width, 0, *img, resize.Lanczos3)
		f, err := w.Create(name)
		if err != nil {
			log.Fatal(err)
		}

		err = png.Encode(f, m)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func getFolderName() string {
	t := time.Now()
	timeStr := t.Format("2006-01-02 15.04.05")
	return fmt.Sprintf("appiconizer %s", timeStr)
}

//called only for normal icon creation.
func resizeImage(width uint, img image.Image, path string) {
	name := fmt.Sprintf("icon_%d.png", width)
	newPath := filepath.Join(path, name)
	m := resize.Resize(width, 0, img, resize.Lanczos3)

	out, err := os.Create(newPath)
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
var Version = "0.1"

const examples = `
examples:
  appiconizer create -source icon.png
`

type options struct {
	file   string
	device string
	target string
	zip    bool
	name   string
}

type command struct {
	fs *flag.FlagSet
	fn func(args []string) error
}
