GOOS=linux 
GOARCH=amd64 
CGO_ENABLED=0 

install: test-deps build-deps ## install from the current working tree
	@npm i
	@go install -v .

clean: ## remove build artifacts from the working tree
	@rm -rf public

test-deps: ## install test dependencies
	@go install github.com/bokwoon95/wgo@latest
	@go get -v ./...
	@go mod tidy

build-deps: ## install build dependencies
	@go install github.com/a-h/templ/cmd/templ@latest

deps: build-deps test-deps ## install build and test dependencies

assets: clean
	@cp -a assets/. ./public/

generate: ## Go and Templ generate
	@npx tailwindcss -i style.css -o ./public/style.css
	@templ generate
	
run: ## run and watch
	@wgo -file=.go -file=.templ -file=.css -file=.md -xfile=_templ.go make generate :: go run main.go --dev

build: assets generate ## build public static
	@go run main.go

update: ## update dependencies
	@go get -u
	@go mod tidy
	@make install
	@go mod tidy

help: ## Show this help.
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v grep | sed -e 's/\\$$//' | sed -e 's/:.\+##/\n\t/'
