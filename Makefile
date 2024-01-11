# Include variables from the .envrc file
include .envrc
###################### HELPERS ###################################
## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

############################### DEV #################################
## run/web: run the cmd/web application
.PHONY: run/web
run/web:
	go run ./cmd/web -db-dsn=${DATABASE_URL} \
	-s3_bucket=${S3_BUCKET} -s3_region=${S3_REGION} -s3_endpoint=${S3_ENDPOINT} \
	-s3_akid=${S3_ACCESS_KEY_ID} -s3_sak=${S3_SECRET_ACCESS_KEY} \
	-env=${ENV} -port=${WEB_PORT} -cdn_host=${CDN_HOST}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${DATABASE_URL}

###################### AUDIT #################################
## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	# @echo 'Formatting code...'
	# go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	# staticcheck ./...
	# @echo 'Running tests...'
	# go test -race -vet=off ./...

####################### BUILD #################################
## build/web: build the cmd/web application
.PHONY: build/web
build/web:
	@echo 'Building cmd/web...'
	go build -ldflags='-s' -o=./bin/web ./cmd/web
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/web ./cmd/web
	GOOS=linux GOARCH=arm64 go build -ldflags='-s' -o=./bin/linux_arm64/web ./cmd/web

## container/build: build the podman image of the application
.PHONY: container/build
container/build:
	@echo 'Building podman image of the web...'
	podman build -t localhost/web_galeria -f Dockerfile .

########################### RUN ##################################
## container/pod/create: create a pod for the web container
.PHONY: container/pod/create
container/pod/create:
	@echo 'Creating the pod web_galeria...'
	podman pod create --name webgaleria --network galeria_net -p ${WEB_PORT}:${WEB_PORT}

## docker/run/web: run a podman container using the build podman image
.PHONY: container/run/web
container/run/web:
	@echo 'Running a container of the web Application'
	podman run -d --pod webgaleria --restart=unless-stopped \
	--name web_galeria localhost/web_galeria -db-dsn=${DATABASE_URL} \
	-s3_bucket=${S3_BUCKET} -s3_region=${S3_REGION} -s3_endpoint=${S3_ENDPOINT} \
	-s3_akid=${S3_ACCESS_KEY_ID} -s3_sak=${S3_SECRET_ACCESS_KEY} \
	-env=${ENV} -port=${WEB_PORT}

################ DEPLOY #################################
## web_galeria/build: build images/network/volume/pod neccesary for galeria app
.PHONY: web_galeria/build
web_galeria/build:
	@echo 'Building all necesary for galeria, this can take a while...'
	make container/build

## web_galeria/deploy: build the app image, create the app pod and run inside all containers related with the galeria app
.PHONY: web_galeria/deploy
web_galeria/deploy:
	@echo 'Deploying the app web_galeria, this can take a while...'
	make container/pod/create
	make container/run/web
