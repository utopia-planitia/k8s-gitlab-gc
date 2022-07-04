# read env vars
include .env
export

.PHONY: up
up:
	d8s up tilt up

.PHONY: down
down:
	d8s run tilt down
	d8s down

.PHONY: tup
tup:
	tilt up

.PHONY: tdown
tdown:
	tilt down