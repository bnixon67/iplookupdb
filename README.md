# iplookupdb

iplookupdb utilizes the MaxMind GeoIP and GeoLite databases to look up IP addresses. To use this tool, you must first create an account and download the necessary database from MaxMind. See [GeoLite2 Free Geolocation Data
](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data).

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
