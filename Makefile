.SILENT:
.PHONY: all tools client server tidy shell

DOCKER_IMAGE:="golang:1.22-bookworm"
CONTAINER_NAME:="golang_builder"

all : tools client server
	@echo "Building all components..."

tools:
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh -c 'git config --global --add safe.directory /srv && scripts/build.tools.sh'

client:
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh -c 'git config --global --add safe.directory /srv && scripts/build.client.sh'

server:
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh -c 'apt update && apt install -yq npm && git config --global --add safe.directory /srv && scripts/build.server.sh'


tidy:
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} go mod tidy

shell:
	@sudo docker run -it --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh

