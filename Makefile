.SILENT:
.PHONY: build tidy shell

DOCKER_IMAGE:="golang:1.22-bookworm"
CONTAINER_NAME:="golang_builder"

build:
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh -c 'git config --global --add safe.directory /srv && scripts/build.client.sh'
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh -c 'git config --global --add safe.directory /srv && scripts/build.server.sh'
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh -c 'git config --global --add safe.directory /srv && scripts/build.tools.sh'

tidy:
	@sudo docker run --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} go mod tidy

shell:
	@sudo docker run -it --rm --name ${CONTAINER_NAME} -v "${PWD}":/srv -w /srv ${DOCKER_IMAGE} sh

