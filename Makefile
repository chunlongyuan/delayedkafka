
LOCALIP=$(shell ./getip.sh)

.PHONY: test
test:
	go test ./... --count=1 -p=1

.PHONY: docker
docker:
	docker build . -t kdqueue:latest

docker_compose_files=-f docker-compose/docker-compose.yml

up:
	docker-compose $(docker_compose_files) up -d

down:
	docker-compose $(docker_compose_files) down --remove-orphans


ps:
	docker-compose $(docker_compose_files) ps

restart:
	$(MAKE) down up

clean:
	docker system prune -f