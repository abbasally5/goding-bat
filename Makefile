
export BUILD_ROOT=${PWD}

GIT_SHA=$(shell git rev-parse --verify HEAd || echo "unknown")

.PHONY: clean
clean:
	@echo "Cleaning"
	rm -rf build/

docker:
	@echo "Building docker image for goding-bat"
	cp deploy/docker/godingbat-service/Dockerfile ${BUILD_ROOT}/build/goding-bat
	@cd ${BUILD_ROOT}/build/goding-bat && docker build --platform linux/amd64 --rm -t godingbat:${GIT_SHA} .

.PHONY: build
build:
	@echo "Building goding-bat binary"
	mkdir -p ${BUILD_ROOT}/build/goding-bat
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ${BUILD_ROOT}/build/goding-bat/godingbat-service github.com/abbasally5/goding-bat/cmd

test:
	@echo "Running tests"
