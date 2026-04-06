.SILENT:
.PHONY: all tools client server web tidy shell

DOCKER_IMAGE:="golang:1.22-bookworm"
CONTAINER_NAME:="golang_builder"

all : tools client server
	@echo "Building all components..."

tools:
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh -c 'git config --global --add safe.directory /srv && scripts/build.tools.sh'

client:
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh -c 'git config --global --add safe.directory /srv && scripts/build.client.sh'

web:
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh -c 'go install github.com/rakyll/statik@latest && $$(go env GOPATH)/bin/statik -m -src="./web-vue3" -f -dest="./server/embed" -p web -ns web'

server: web
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh -c 'git config --global --add safe.directory /srv && scripts/build.server.sh'


tidy:
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} go mod tidy

shell:
	@sudo docker run -it --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh

