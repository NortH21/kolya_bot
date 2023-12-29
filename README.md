<!-- TOC -->

<!-- /TOC -->

<a target="_blank" href="https://hub.docker.com/r/north21/kolya_bot"><img src="https://img.shields.io/docker/pulls/north21/kolya_bot" /></a>
<a target="_blank" href="https://hub.docker.com/r/north21/kolya_bot/tags"><img src="https://img.shields.io/docker/v/north21/kolya_bot/latest?label=docker%20image%20ver." /></a>
<a target="_blank" href="https://github.com/NortH21/kolya_bot/graphs/commit-activity"><img src="https://img.shields.io/github/last-commit/NortH21/kolya_bot" /></a>
<a target="_blank" href="https://github.com/NortH21/kolya_bot/graphs/contributors"><img src="https://img.shields.io/github/contributors/NortH21/kolya_bot" /></a>
<a target="_blank" href="https://github.com/NortH21/kolya_bot/issues"><img src="https://img.shields.io/github/issues/NortH21/kolya_bot" /></a>
<a target="_blank" href="https://github.com/NortH21/kolya_bot/actions"><img src="https://img.shields.io/github/actions/workflow/status/NortH21/kolya_bot/docker-publish.yml"  /></a>

# Бот Колян
Может послать или нагрубить.

## Usage
If necessary, the linux/amd64 platform has been changed.
### Docker
```
docker build --platform linux/amd64 -t kolya_bot:latest .
docker run -d --name kolya_bot -e TELEGRAM_APITOKEN='token' kolya_bot:latest
```

### OR

### Docker-compose
Create .env file and run:
```
docker-compose up --build
```