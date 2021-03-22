build:
	@echo "Building docker image"
	docker build --pull -f ./Dockerfile -t hellofresh .

help:
	@echo "Show help"
	docker run --rm hellofresh -h