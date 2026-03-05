RESTLESS: MAXIMUM TERMINAL VIOLENCE (Cinematic CLI Demo)
=======================================================

This is a non-commit demo pack (Dockerfile + scripts) intended to be dropped
into the root of your Restless repo and run locally.

Files:
- Dockerfile
- demo.sh      (interactive cyberpunk cinematic demo)
- run.sh       (optional asciinema recording wrapper)

Build:
  docker build -t restless-violence .

Run:
  docker run --rm -it -e TERM=xterm-256color restless-violence

Silence beeps:
  docker run --rm -it -e TERM=xterm-256color -e NO_BEEP=1 restless-violence

Record (asciinema):
  docker run --rm -it -e TERM=xterm-256color -e RECORD=1 restless-violence

Extract recording:
  docker ps -a
  docker cp <container_id>:/demo/out/session.cast .
  asciinema play session.cast
