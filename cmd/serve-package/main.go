package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/coreos/go-omaha/omaha"
)

func main() {
	pkgfile := flag.String("package-file", "", "Path to the update payload")
	version := flag.String("package-version", "", "Semantic version of the package provided")
	listenAddress := flag.String("listen-address", ":8000", "Host and IP to listen on")

	flag.Parse()

	if *pkgfile == "" {
		fmt.Println("package-file is a required flag")
		os.Exit(1)
	}

	if *version == "" {
		fmt.Println("package-version is a required flag")
		os.Exit(1)
	}

	server, err := omaha.NewTrivialServer(*listenAddress)
	if err != nil {
		fmt.Printf("failed to make new server: %v\n", err)
		os.Exit(1)
	}

	server.SetVersion(*version)
	err = server.AddPackage(*pkgfile, "update.gz")
	if err != nil {
		fmt.Printf("failed to add package: %v\n", err)
		os.Exit(1)
	}

	err = server.Serve()
	if err != nil {
		fmt.Printf("server exited with an error: %v\n", err)
		os.Exit(1)
	}
}
