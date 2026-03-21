build:
	go build -o anytype

test:
	go test -v ./tests/...

sunbeam-install:
	sunbeam extension install ./anytype

raycast-install:
	cp anytype ~/Library/Application\ Support/com.raycast.macos/scripts/anytype

hyper-install:
	@echo "Installing hyper-sunbeam plugin..."
	hyper i hyper-sunbeam
	@echo "Updating ~/.hyper.js configuration..."
	@if grep -q "hyper-sunbeam" ~/.hyper.js 2>/dev/null; then \
		echo "Plugin already configured in ~/.hyper.js"; \
	else \
		node -e " \
			const fs = require('fs'); \
			let config; \
			try { config = require(process.env.HOME + '/.hyper.js'); } \
			catch(e) { config = { config: {}, plugins: [] }; } \
			if (!config.plugins) config.plugins = []; \
			if (!config.plugins.includes('hyper-sunbeam')) { \
				config.plugins.push('hyper-sunbeam'); \
			} \
			if (!config.config) config.config = {}; \
			if (!config.config.hyperSunbeam) { \
				config.config.hyperSunbeam = { hotkey: 'Ctrl+;' }; \
			} \
			const content = 'module.exports = ' + JSON.stringify(config, null, 2) + ';\n'; \
			fs.writeFileSync(process.env.HOME + '/.hyper.js', content); \
			console.log('Configuration updated successfully!'); \
		"; \
	fi
	@echo "Hyper plugin installed. Restart Hyper to apply changes."

create-tag:
	@read -p "Enter release version: " release_version; \
	git tag -a v$$release_version -m "Release version $$release_version"; \
	git push --tags