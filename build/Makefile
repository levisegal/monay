.PHONY: build build.nocache build.db build.api build.dashboard build.service build.service.nocache up up.dashboard down down.clean logs db.reinit

build:
	docker compose \
		--profile database \
		--profile monay-apis \
		build

build.nocache:
	docker compose \
		--profile database \
		--profile monay-apis \
		build --no-cache

build.db:
	docker compose --profile database build postgres

build.api:
	docker compose --profile monay-apis build holdings

build.dashboard:
	docker compose --profile dashboard build dashboard

# Build a single service: make build.service SERVICE=holdings
build.service:
	docker compose --profile database --profile monay-apis build $(SERVICE)

# Build a single service without cache: make build.service.nocache SERVICE=holdings
build.service.nocache:
	docker compose --profile database --profile monay-apis build --no-cache $(SERVICE)

up:
	docker compose \
		--profile database \
		--profile monay-apis \
		up

up.dashboard:
	docker compose --profile dashboard up

down:
	docker compose \
		--profile database \
		--profile monay-apis \
		down

down.clean:
	docker compose \
		--profile database \
		--profile monay-apis \
		down --volumes --remove-orphans

logs:
	docker compose logs -f

db.reinit:
	./postgres/reinit.sh
