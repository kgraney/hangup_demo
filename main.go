package main

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"

	"github.com/urfave/cli"
)

func generateRandomPages(numPages int, pageSize int) map[string][]byte {
	bytes := make(chan []byte)
	go func() {
		for i := 0; i < numPages; i++ {
			buf := make([]byte, pageSize)
			rand.Read(buf)
			bytes <- buf
		}
		close(bytes)
	}()

	pages := make(map[string][]byte)
	for buf := range bytes {
		hash := fmt.Sprintf("%x", md5.Sum(buf))
		pages[hash] = buf
	}
	return pages
}

func runServer(c *cli.Context) error {
	pages := generateRandomPages(100, 2e3)

	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		for k := range pages {
			fmt.Fprintf(w, "%s\n", k)
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		k := r.URL.Path[1:]
		fmt.Fprintf(w, "%s", pages[k])
	})
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
