package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	currentDir, _ = os.Getwd()
	mode          string
	size          int64
)

type Image struct {
	path   string
	config image.Config
}

func main() {
	setFlag()
	if _, err := exec.Command("ls", currentDir+"/.git").Output(); err != nil {
		fmt.Fprintf(os.Stderr, "err: %s, not found .git dir.\n", err)
		os.Exit(1)
	}

	git := GetGitCommandPath()

	out, err := exec.Command(git, "ls-files", "--other", "--modified", "--cached").Output()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	res := string(out)
	line := strings.Split(res, "\n")

	var imgList []Image
	for _, path := range line {
		r := regexp.MustCompile(`.+\.(jpg|png|gif)`).MatchString(path)
		if r {
			f, err := os.Open(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "faild open file. "+path)
				continue
			}
			fileinfo, err := f.Stat()
			if fileinfo.Size() < 1 || err != nil {
				fmt.Fprintln(os.Stderr, "file size zero. "+path)
				continue
			}
			if size > 0 {
				if fileinfo.Size() > size {
					fmt.Fprintf(os.Stderr, "file size over the limit %d byte. path: %s \n",
						size, path)
					continue
				}
			}
			img := Image{path, GetImageSize(path)}
			fmt.Printf("%s, width: %d, height: %d\n",
				img.path, img.config.Width, img.config.Height)
			imgList = append(imgList, img)
		}
	}

	for _, img := range imgList {
		if mode == "mul4" {
			if img.config.Width%4 != 0 ||
				img.config.Height%4 != 0 {
				fmt.Fprintf(os.Stderr,
					"Bad file size. not multiple of 4. %s width: %d, height: %d\n",
					img.path, img.config.Width, img.config.Height)
			}
		} else {
			if img.config.Width&(img.config.Width-1) != 0 ||
				img.config.Height&(img.config.Height-1) != 0 {
				fmt.Fprintf(os.Stderr,
					"Bad file size. not power of 2. %s width: %d, height: %d\n",
					img.path, img.config.Width, img.config.Height)
			}
		}
	}
}

func setFlag() {
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&mode, "mode", "pow2", "pow2 / mul4")
	f.Int64Var(&size, "size", 0, "data size limit")
	f.Parse(os.Args[1:])
	for 0 < f.NArg() {
		f.Parse(f.Args()[1:])
	}
}

func GetGitCommandPath() string {
	out, err := exec.Command("which", "git").Output()
	if err != nil {
		if binary.Size(out) == 0 {
			out, _ = exec.Command("whereis", "git").Output()
			if binary.Size(out) == 0 {
				fmt.Fprint(os.Stderr, "not found git command path.\n")
			}
		}
	}
	path := strings.Trim(string(out), "\n")
	return path
}

func GetImageSize(path string) image.Config {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return img
}
