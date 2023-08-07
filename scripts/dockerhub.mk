DOCKER_REPOSITORY = bluenviron/mavp2p

define DOCKERFILE_DOCKERHUB
FROM scratch
ARG TARGETPLATFORM
ADD tmp/binaries/$$TARGETPLATFORM.tar.gz /
ENTRYPOINT [ "/mavp2p" ]
endef
export DOCKERFILE_DOCKERHUB

dockerhub:
	$(eval VERSION := $(shell git describe --tags | tr -d v))

	docker login -u $(DOCKER_USER) -p $(DOCKER_PASSWORD)

	rm -rf tmp
	mkdir -p tmp tmp/binaries/linux/arm

	cp binaries/*linux_amd64.tar.gz tmp/binaries/linux/amd64.tar.gz
	cp binaries/*linux_armv6.tar.gz tmp/binaries/linux/arm/v6.tar.gz
	cp binaries/*linux_armv7.tar.gz tmp/binaries/linux/arm/v7.tar.gz
	cp binaries/*linux_arm64v8.tar.gz tmp/binaries/linux/arm64.tar.gz

	docker buildx rm builder 2>/dev/null || true
	rm -rf $$HOME/.docker/manifests/*
	docker buildx create --name=builder --use

	echo "$$DOCKERFILE_DOCKERHUB" | docker buildx build . -f - \
	--provenance=false \
	--platform=linux/amd64,linux/arm/v6,linux/arm/v7,linux/arm64/v8 \
	-t $(DOCKER_REPOSITORY):$(VERSION) \
	-t $(DOCKER_REPOSITORY):latest \
	--push

	docker buildx rm builder
	rm -rf $$HOME/.docker/manifests/*
