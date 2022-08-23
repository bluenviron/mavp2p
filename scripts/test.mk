test:
	docker run --rm -v $(PWD):/app -w /app \
	$(BASE_IMAGE) \
	go build -o /dev/null .
