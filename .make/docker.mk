GETDOCKERFILE=$(firstword $(subst ~, ,$1))
GETDOCKERIMAGE=$(or $(word 2,$(subst ~, ,$1)),$(value 2))

.PHONY=docker docker/% docker.push docker.push/%
docker: prebuild $(DOCKER_IMAGES:%=docker/%) ## build docker image
docker/%:
	@echo Build image $(DOCKER_REGISTRY)/$(call GETDOCKERIMAGE,$(@:docker/%=%)):$(BUILDNUMBER) from docker/$(call GETDOCKERFILE,$(@:docker/%=%))
	docker build -t $(DOCKER_REGISTRY)/$(call GETDOCKERIMAGE,$(@:docker/%=%)):$(BUILDNUMBER) -f docker/$(call GETDOCKERFILE,$(@:docker/%=%)) .

docker.push: $(DOCKER_IMAGES:%=docker.push/%) ## push docker images to the registry
docker.push/%:
	@echo Push docker image $(DOCKER_REGISTRY)/$(call GETDOCKERIMAGE,$@):$(BUILDNUMBER)
	docker push $(DOCKER_REGISTRY)/$(call GETDOCKERIMAGE,$@):$(BUILDNUMBER)
