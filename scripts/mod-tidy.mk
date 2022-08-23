mod-tidy:
	docker run --rm -it -v $(PWD):/s -w /s $(BASE_IMAGE) \
	sh -c "go get && go mod tidy"
