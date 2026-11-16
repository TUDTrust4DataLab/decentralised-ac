.PHONY: run-dev
run-dev:
	@echo "Running CIPHER Fabric Gateway"
	@npm run dev

.PHONY: build
build:
	@echo "Building CIPHER Fabric Gateway"
	@npm run build
	@echo "Done!"

.PHONY: clean
clean:
	@echo "Cleaning up CIPHER Fabric Gateway dist and node_modules"
	@rm -rf dist node_modules
	@echo "Done!"

.PHONY: install
install:
	@echo "Installing CIPHER Fabric Gateway node modules"
	@make clean
	@npm install
	@echo "Done!"
