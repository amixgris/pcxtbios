/*
Copyright (C) 2019-2020 Andreas T Jonsson

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

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
