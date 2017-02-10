package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli"
)

func generateRandomPage(pageSize int) (string, string) {
	log.Printf("Generating random response page of size %d bytes", pageSize)
	buf := make([]byte, pageSize)
	rand.Read(buf)

	hash := fmt.Sprintf("%x", md5.Sum(buf))
	return hash, base64.StdEncoding.EncodeToString(buf)
}

func runServer(c *cli.Context) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hash, bstr := generateRandomPage(2e3)

		fmt.Fprintf(w, "hash: %s\n\n%s", hash, bstr)
		//k := r.URL.Path[1:]
	})

	log.Print("Starting listening server")
	return http.ListenAndServe(":8080", nil)
}

func sendRequests(c *cli.Context) error {
	resp, err := http.Get("http://localhost:8080/list")
	if err != nil {
		return err
	}
	fmt.Println(resp)
	return nil
}

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "kerberos",
			Usage: "Use the Kerberos proxy HTTP transport",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "request",
			Usage:  "Send requests to local server",
			Action: sendRequests,
		},
		{
			Name:   "server",
			Usage:  "Start a local random page server",
			Action: runServer,
		},
	}

	app.Run(os.Args)
}
