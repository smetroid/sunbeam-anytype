build:
	go build -o sunbeam-anytype

sunbeam-install:
	sunbeam extension install anytype.sh

create-tag:
	@read -p "Enter release version: " release_version; \
	git tag -a v$$release_version -m "Release version $$release_version"; \
	git push --tags
