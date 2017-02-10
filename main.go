package main

import (
	"bufio"
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

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
	certFile := c.String("cert")
	keyFile := c.String("key")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request for URL: /%s", r.URL.Path[1:])
		hash, bstr := generateRandomPage(184643)

		fmt.Fprintf(w, "hash: %s\n\n%s", hash, bstr)
	})

	log.Print("Starting listening server")
	return http.ListenAndServeTLS(":8080", certFile, keyFile, nil)
}

func sendRequests(c *cli.Context) error {
	server := c.String("server")
	log.Printf("Connecting to target server: %s", server)

	urls := make(chan string)
	errs := make(chan error)

	getThread := func() {
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Timeout:   20 * time.Second,
			Transport: transport,
		}

		for url := range urls {
			fmt.Printf("Fetching URL %s\n", url)
			resp, err := client.Get(url)
			if err != nil {
				errs <- err
			}

			reader := bufio.NewReader(resp.Body)
			var hash string
			for {
				line, err := reader.ReadBytes('\n')
				if err != nil {
					log.Printf("%s\n", err)
					break
				}
				if string(line) == "\n" {
					continue
				} else if string(line[0:4]) == "hash" {
					hash = string(line[6 : len(line)-1])
				}
			}
			fmt.Println(hash)
		}
	}

	for i := 0; i < 10; i++ {
		go getThread()
	}

	for {
		_, urlRand := generateRandomPage(30)
		url := fmt.Sprintf("https://%s/%s", server, urlRand)
		urls <- url
	}

	err := <-errs
	return err
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
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "cert",
					Usage: "The SSL certificate file",
				},
				cli.StringFlag{
					Name:  "key",
					Usage: "The SSL key file",
				},
			},
		},
	}

	app.Run(os.Args)
}
