.PHONY: docker
docker:
	@rm webook || true
	@GOOS=linux GOARCH=arm go build -o webook .
	@docker rmi -f lalalalade/webook-live:v0.0.1
	@docker build -t lalalalade/webook-live:v0.0.1 .
