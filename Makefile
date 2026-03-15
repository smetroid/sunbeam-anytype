build:
	go build -o sunbeam-anytype

sunbeam-install:
	sunbeam extension install anytype.sh

raycast-ext-install:
	cd extension && npm install
	cd extension && npx ray build -e dist -o ./dist
	rm -rf ~/.local/share/raycast/Extensions/sunbeam-anytype
	mkdir -p ~/.local/share/raycast/Extensions/sunbeam-anytype
	cp -r extension/dist/* ~/.local/share/raycast/Extensions/sunbeam-anytype/
	cp extension/package.json ~/.local/share/raycast/Extensions/sunbeam-anytype/
	@echo "Extension installed to ~/.local/share/raycast/Extensions/sunbeam-anytype/"
	@echo "Restart Raycast to use the extension"

raycast-ext-dev:
	cd extension && npm install
	cd extension && npx ray develop

create-tag:
	@read -p "Enter release version: " release_version; \
	git tag -a v$$release_version -m "Release version $$release_version"; \
	git push --tags
