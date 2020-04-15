# cdn_version_scraper

## A version scraping tool for automated OpenTTD builds

_cdn\_version\_scraper_ is a super simple utility that does the following:

1. Pulls and parses the OpenTTD CDN manifest
2. Determines the most recent available version given a desired stability - and automatically selects a more stable version if one is available
3. Returns a ENVVAR string with the correct target version, ready to be used in your scripts

## Running

```
go get github.com/ropenttd/cdn_version_scraper
go run github.com/ropenttd/cdn_version_scraper stable
```

### Use in CI

Instead of passing the channel via the command line, you can instead pass it as an environment variable:

```
export PLUGIN_CHANNEL=stable
go run github.com/ropenttd/cdn_version_scraper
```

Additionally, setting `PLUGIN_OUTPUTFILE` will additionally write a file containing the output to the specified path.

These environment variables make using this perfect for use as a Drone CI plugin - see the main build repository for more information.