lint: lint-static lint-templates

format: format-go format-static format-templates

format-go:
	go fmt ./...

lint-go:
	golangci-lint run

lint-static:
	npx prettier --check "static/**/*.{js,css}"

format-static:
	npx prettier "static/**/*.{js,css}" --write

lint-templates:
	djlint templates --lint

format-templates:
	djlint templates/*.html --reformat