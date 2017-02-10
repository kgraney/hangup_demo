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
		hash, bstr := generateRandomPage(184643)

		fmt.Fprintf(w, "hash: %s\n\n%s", hash, bstr)
		//k := r.URL.Path[1:]
	})

	log.Print("Starting listening server")
	return http.ListenAndServe(":8080", nil)
}

func sendRequests(c *cli.Context) error {
	server := c.String("server")
	log.Printf("Connecting to target server: %s", server)

	_, urlRand := generateRandomPage(30)
	url := fmt.Sprintf("http://%s/%s", server, urlRand)

	for {
		fmt.Printf("Fetching URL %s\n", url)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		fmt.Println(resp)
	}
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
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "server",
					Usage: "The host:port string for the server to connect to",
				},
			},
		},
		{
			Name:   "server",
			Usage:  "Start a local random page server",
			Action: runServer,
		},
	}

	app.Run(os.Args)
}
