lint: lint-go  lint-templates lint-static

format: format-go format-static format-templates

format-go:
	npm run format-go

lint-go:
	npm run lint-go

lint-static:
	npm run lint-static

format-static:
	npm run format-static

lint-templates:
	npm run lint-templates

format-templates:
	npm run format-templates
