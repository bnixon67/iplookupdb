// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

/*
iplookupdb utilizes the MaxMind GeoIP and GeoLite databases to look up IP
addresses. To use this tool, you must first create an account and download the
necessary database from MaxMind.
See https://dev.maxmind.com/geoip/geolite2-free-geolocation-data.

Usage:

  iplookupdb [flags] [ip address ...]

The flags are:

  -db string
    	Path to the GeoLite2 City database (default "GeoLite2-City.mmdb")
  -delimiter string
    	Delimiter for the CSV output. (default ",")
  -in string
    	Input file path. If not specified, reads from standard input.
  -lang string
    	Language for GeoIP lookup results. (default "en")
  -out string
    	Output file path. If not specified, writes to standard output.

You can specify IP addresses directly via the command line. Use the -in flag
to read from a file. If no IP addresses are provided on the command line and
the -in flag is not used, the program reads from stdin.

The output is a comma-separated list of the IP address, city, subdivision
(e.g., state for US-based addresses), and country. To change the separator,
use the -delimiter flag. By default, the output is sent to stdout unless
the -out flag is specified.

*/

package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// config contains the command-line flags.
type config struct {
	dbName     string
	inputName  string
	outputName string
	lang       string
	delimiter  rune
}

// parseFlags parses and does some simple validation of the command-line flags.
func parseFlags() (config, error) {
	dbName := flag.String("db", "GeoLite2-City.mmdb", "Path to the GeoLite2 City database")
	inputFile := flag.String("in", "", "Input file path. If not specified, reads from stdin.")
	outputFile := flag.String("out", "", "Output file path. If not specified, writes to stdout.")
	lang := flag.String("lang", "en", "Language for GeoIP lookup results.")
	delimiter := flag.String("delimiter", ",", "Delimiter for the CSV output.")
	flag.Parse()

	if len(flag.Args()) > 0 && *inputFile != "" {
		return config{}, errors.New("cannot provide both -in and IPs on command line")
	}

	if len(*delimiter) != 1 {
		return config{}, errors.New("must specify a single character as a delimiter")
	}
	delimRune := rune((*delimiter)[0])

	return config{*dbName, *inputFile, *outputFile, *lang, delimRune}, nil
}

// openInput returns an io.ReadCloser based on the name.
// If name is empty, then stdin is used.
func openInput(name string) (io.ReadCloser, error) {
	if name != "" {
		return os.Open(name)
	}
	return os.Stdin, nil
}

// openOutput returns an io.WriteCloser based on the name.
// If name is empty, then stdout is used.
// If name is provided, the file must not exist, otherwise an error is returned.
func openOutput(name string) (io.WriteCloser, error) {
	if name != "" {
		flag := os.O_WRONLY | os.O_CREATE | os.O_EXCL
		return os.OpenFile(name, flag, 0666)
	}
	return os.Stdout, nil
}

// processIP will lookup the ipStr provided in db and output the results to w.
//
// The output is a comma-separated list of IP Address, city, subdivision
// (e.g., state for US-based addresses), and county.
//
// If the IP is private, then "private" is returned for city, subdivision,
// and county.
//
// If city, subdivision, or county is empty, then unknown is used.
//
// Any errors are displayed on stderr, such as parsing or searching fails.
func processIP(w *csv.Writer, db *geoip2.Reader, ipStr, lang string) {
	ipStr = strings.TrimSpace(ipStr)
	ip := net.ParseIP(ipStr)
	if ip == nil {
		fmt.Fprintf(os.Stderr, "Cannot convert %q to IP\n", ipStr)
		return
	}

	record, err := db.City(ip)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error for IP %v: %v\n", ip, err)
		return
	}

	cityName, subName, countryName := "", "", ""
	if len(record.Subdivisions) > 0 {
		subName = record.Subdivisions[0].Names[lang]
	}
	cityName = record.City.Names[lang]
	countryName = record.Country.Names[lang]

	if ip.IsPrivate() {
		cityName = "private"
		subName = "private"
		countryName = "private"
	}

	fields := []string{ip.String(), cityName, subName, countryName}
	for n := range fields {
		if fields[n] == "" {
			fields[n] = "unknown"
		}
	}

	// write and flush immediately for interactive use
	w.Write(fields)
	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Fprintln(os.Stderr, "error writing csv:", err)
	}
}

func processIPsFromArgs(args []string, db *geoip2.Reader, w *csv.Writer, lang string) {
	for index := range args {
		processIP(w, db, args[index], lang)
	}
}

func processIPsFromInput(r io.ReadCloser, db *geoip2.Reader, w *csv.Writer, lang string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		processIP(w, db, scanner.Text(), lang)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func main() {
	cfg, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid option: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	db, err := geoip2.Open(cfg.dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(2)
	}
	defer db.Close()

	input, err := openInput(cfg.inputName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open input: %v\n", err)
		os.Exit(3)
	}
	defer input.Close()

	output, err := openOutput(cfg.outputName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open output: %v\n", err)
		os.Exit(4)
	}
	defer output.Close()

	csvWriter := csv.NewWriter(output)
	csvWriter.Comma = cfg.delimiter

	args := flag.Args()
	if len(args) > 0 {
		processIPsFromArgs(args, db, csvWriter, cfg.lang)

	} else {
		if cfg.inputName == "" {
			fmt.Printf("Please provide IPs, one per line:\n")
		}

		processIPsFromInput(input, db, csvWriter, cfg.lang)
	}
}
