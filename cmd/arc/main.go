package main

import (
	"flag"
	"log"

	"github.com/jm33-m0/arc"
)

func main() {
	tarball := flag.String("f", "", "Archive to extract")
	dst := flag.String("d", "", "Destination directory")
	flag.Parse()

	if *tarball == "" || *dst == "" {
		log.Fatal("Both archive and destination must be specified")
	}

	err := arc.Unarchive(*tarball, *dst)
	if err != nil {
		log.Fatal(err)
	}
}
