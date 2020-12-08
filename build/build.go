package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/andreas-jonsson/virtualxt/version"
)

var outputFile = "../../vxtbios.bin"

func init() {
	flag.StringVar(&outputFile, "out", outputFile, "")
}

func main() {
	flag.Parse()
	log.Print("Building...")

	script := "make_linux.sh"
	if runtime.GOOS == "windows" {
		script = "make_win.bat"
	}

	dir, _ := os.Getwd()
	dir = filepath.Join(dir, "..")
	log.Print("Working dir: ", dir)

	var out bytes.Buffer
	cmd := exec.Command(filepath.Join(dir, script))
	cmd.Dir = dir
	cmd.Stderr = &out

	if cmd.Run() != nil {
		log.Print("Could not execute command: ", out.String())
		os.Exit(-1)
	}

	data, err := ioutil.ReadFile(filepath.Join(dir, "pcxtbios.bin"))
	if err != nil {
		log.Print(err)
		os.Exit(-1)
	}

	ver := version.Current.String()
	log.Print("Versin is: ", ver)

	oldStr, newStr := []byte("VirtualXT BIOS v?.?    "), []byte("VirtualXT BIOS v"+ver)

	diff := len(oldStr) - len(newStr)
	for i := 0; i < diff; i++ {
		newStr = append(newStr, 0x20)
	}
	data = bytes.Replace(data, oldStr, newStr, 1)

	fp, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Print(err)
		os.Exit(-1)
	}
	defer fp.Close()

	size, err := io.Copy(fp, bytes.NewReader(data))
	if err != nil {
		log.Print(err)
		os.Exit(-1)
	}

	if _, err := fp.Seek(0, io.SeekStart); err != nil {
		log.Print(err)
		os.Exit(-1)
	}

	var (
		sum byte
		buf [1]byte
	)

	for i := 0; i < int(size-1); i++ {
		if _, err := fp.Read(buf[:]); err != nil {
			log.Print(err)
			os.Exit(-1)
		}
		sum += buf[0]
	}

	buf[0] = byte(256 - int(sum))
	if _, err := fp.WriteAt(buf[:], size-1); err != nil {
		log.Print(err)
		os.Exit(-1)
	}

	log.Printf("Checksum: 0x%X", buf[0])
}
