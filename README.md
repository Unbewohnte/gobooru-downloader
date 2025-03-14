# GOBOORU-DOWNLOADER
## Automatic Booru media downloader

helps to bulk download booru media with tags.

Features:
- media download with specific tags
- start at any page
- custom download limit
- custom max file size limit
- ability to download only images/only video
- http/socks5 proxy support
- custom worker count
- request retry system

Boorus supported:
- danbooru.donmai.us
- gelbooru.com


In the course of running the program, metadata is saved alongside the content. Currently (might be outdated) metadata files have the same SHA hash names as images they belong to with the suffix of `_metadata.json` and the structure is as follows:

```json
{
  "tags": [
    "general",
    "tags",
    "go",
    "here"
  ],
  "copyright": [
    "genshin_impact",
    "(for example)"
  ],
  "characters": [
    "characters",
    "here"
  ],
  "artists": [
    "artists",
    "here"
  ],
  "hash": "0a1a20ede5a8a3e2c56907f6099b7e0452a5c3730c3338a5dcdd18390fc81534",
  "from_host": "danbooru.donmai.us",
  "url": "https://cdn.donmai.us/original/someImage.png"
}
```

Note that gelbooru does not separate tags so `copyright`, `characters` and `artists` are put alongside general tags inside `tags` field.

## Usage

Use `-help` to output the latest information on flags and their purpose.

`./gobooru-downloader -help`

### Flags

| Flag | Description | Default value |
|:---:|:---:|:---:|
| version | Print version information and exit | false |
| url | URL to the booru page (blank for danbooru.donmai.us) | https://danbooru.donmai.us/ |
| proxy | Set proxy connection string | "" |
| workers | Set worker count | 8 |
| output | Set output directory name | output |
| silent | Output nothing to the console | false |
| max-retries | Set max http request retry count | 3 |
| only-images | Save only images | false |
| only-videos | Save only videos | false |
| tags | Set tags | "" |
| from-page | Set initial page number | 1 |
| max-filesize-mb | Set max file size in megabytes to be allowed for download (0 for no cap) | 0 |
| download-limit-gb | Set download limit in gigabytes. The program will quit after the limit was reached (0 for no cap) | 0.0 |

### Examples


| Command | Description |
|:---:|:---:|
| gobooru-downloader -from-page 3 -only-images -tags "bocchi_the_rock!" | Downloads only images with "bocchi_the_rock!" tag from the 3rd page of danbooru.donmai.us |
| gobooru-downloader -proxy "socks5://127.0.0.1:1080" -url "https://gelbooru.com/" | Downloads everything starting from the first page from gelbooru, requests are issued through specified socks5 proxy |
| gobooru-downloader -only-images -download-limit-gb 20 -output danbooruDownloads | Downloads any image from danbooru.donmai.us to danbooruDownloads directory. Stops after 20 gigabytes of content was downloaded |
| gobooru-downloader -max-retries 6 -max-filesize-mb 5 -tags "rating:g" | Downloads any content smaller than 5 megabytes from danbooru.donmai.us with rating:g, in case of errors, retries 6 times.  |
| gobooru-downloader -proxy "socks5://127.0.0.1:1080" -only-images -download-limit-gb 15 -max-retries 8 -max-filesize-mb 6 -tags "rating:g order:score" -from-page 1 -workers 4 | Downloads images from danbooru.donmai.us of less than 6 megabytes, rating:g and ordered by score, 4 workers are used. Will stop after 15 gigabytes of data had been downloaded. Try using something like this one for long download sessions |


## Build

Run `go build` in the project directory

## License

gobooru-downloader is licensed under GNU Public License v3. See `COPYING` for information