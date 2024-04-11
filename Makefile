lint: lint-go  lint-templates lint-static

format: format-go format-static format-templates

format-go:
	go fmt ./...

lint-go:
	golangci-lint run

lint-static:
	npx prettier --check "static/**/*.{js,css}"
	npx eslint static/javascript/*.js

format-static:
	npx prettier "static/**/*.{js,css}" --write

lint-templates:
	pipenv run djlint templates --lint

format-templates:
	pipenv run djlint templates/*.html --reformat
