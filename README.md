# mangobars

Fast bulk check SSL certificate expiration status for the given servers. The list of servers can be specified
via CSV file for bulk verification or single host
can be specified on the command line. The colorized status is presented on the Console and CSV file (for bulk check).

[![output](https://asciinema.org/a/R5LaiQVf0Ls70e8tth0J1GuCa.svg)](https://asciinema.org/a/R5LaiQVf0Ls70e8tth0J1GuCa)

## Usage

```
Usage: mangobars [OPTION] [FILEPATH]
Checks the expiration status for Server certificates.
Example:
	mangobars -w 20 -a 10 -i host.csv -o result.csv
	mangobars -h amazon.com -p 443
	mangobars -h amazon.com
	mangobars -h amazon.com:443

  -a int
    	Alert if the certificate expiration is less than specified days. (default 10 days)
  -w int
    	Warn if the certificate expiration is less than specified days but has enough time not to be alerted. (default 20)
  -h string
    	Hostname with or without port. Input file specified with `-i` will be ignored.
  -p string
    	Port (default "443")
  -i string
    	CSV file containing host information. (default "host.csv")
  -o string
    	Output file name. (default "result.csv")
  
```
## Examples

### Checking single host
```
$ mangobars -h amazon.com:443
```
Output
```
VALID    amazon.com:443 (*.peg.a2z.com | 117 days | 2021-02-23 12:00:00 +0000 UTC)
```

### Bulk check: Reading host information from CSV file.
```
$ mangobars -w 20 -a 10 -i host.csv -o result.csv
```
Output
```
   VALID    www.packtpub.com:443 (packtpub.com | 246 days | 2021-07-02 12:00:00 +0000 UTC)
   VALID    amplifi.com:443 (*.amplifi.com | 101 days | 2021-02-07 12:00:00 +0000 UTC)
   VALID    www.ui.com:443 (ubnt.com | 70 days | 2021-01-07 12:00:00 +0000 UTC)
   VALID    9to5mac.com:443 (9to5mac.com | 57 days | 2020-12-24 23:04:46 +0000 UTC)
   VALID    thenewstack.io:443 (*.thenewstack.io | 68 days | 2021-01-05 12:59:44 +0000 UTC)
   VALID    9to5google.com:443 (9to5google.com | 53 days | 2020-12-21 01:52:52 +0000 UTC)
   EXPIRED expired.badssl.com:443 (*.badssl.com | -2026 days | 2015-04-12 23:59:59 +0000 UTC)
```
### CSV Format
A host information in the CSV file is expecting the following format.

```
server-host-name, ssl-port
```

### Example
```
amazon.com,443
google.com,443
cloudflare.com,443
microsoft.com,443
```
### Result Format

1. First field indicates the SSL certificate status (`EXPIRED`, `VALID`, `WARN`, `ALERT`). 
    * `EXPIRED` : The server certificate is expired.
    * `VALID` : The server certificate is valid.
    * `WARN` : The server certificate expiration is days away from the specified day (via `-w`)
    * `ALERT` : The server certificate expiring very soon and is only few days away as specified via `-a`.

2. Second field specifies the host and port used for SSL connection.
3. Third field (in paranthesis) specifies the `Subject` value from the X.509 certificate. 
4. Fourth field (in paranthesis) specifies the number of days to expiry from `now`. Negative value means, the certificate has been the days past expiry.
5. Fifth field (in paranthesis) indicates the actual expiration date and time in UTC format.
6. In case of errors, the error string will be specified in the paranthesis. ```ERROR Microsoftonline.com:443 (dial tcp 52.178.167.109:443: i/o timeout)```
