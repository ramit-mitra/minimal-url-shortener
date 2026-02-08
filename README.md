# minimal-url-shortener

## Getting Started

This is my `weekend hobby project` which is a `headless` URL shortener service.

### Tech stack

- [Go](https://go.dev/)
- [Upsun](https://upsun.com/)
- [Docker](https://www.docker.com/), used in local for development

## What it does?

Generate a `short link` for a given URL.

## Routes

| HTTP Method | Path | Description |
| ----------- | ---- | ----------- |
| GET | `/` | default endpoint |
| POST | `/` | create a `short link` |
| GET | `/:code` | send `short code` and get redirected to the actual URL (if exists) |

## Development

To start the development server run:

```bash
docker compose up --build --watch
```

Open <http://localhost:1234/> with your browser to see the result. Prefer using `curl` like a true geek.

## Examples

### Get homepage

```bash
curl -sk http://localhost:1234/ | jq .
```

### Create a shortcode

```bash
curl -skX POST http://localhost:1234/ -d '{"url": "https://ramit.io"}' | jq .
```

### Retrieve URL using shortcode

```bash
curl -sk http://localhost:1234/24qAMU10CBG
```

## Deploy in production

Deploy to [upsun.com](https://upsun.com/). Configuration files are present in `.upsun` directory.

## Suggestions and feedback

Got ideas 💡 about a `feature` or an `enhancement`? Feel free to [open a PR](https://github.com/ramit-mitra/minimal-url-shortener/pulls).

Found a 🐞? Feel free to [open a PR](https://github.com/ramit-mitra/minimal-url-shortener/pulls) and contribute.
