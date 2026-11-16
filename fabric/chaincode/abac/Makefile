.PHONY: test
test:
	@echo 'Testing CIPHER Fabric Chaincode'
	@gofmt -w .
	@go test -v ./...

.PHONY: build
build:
	@echo "Building CIPHER Fabric Chaincode"
	@make test
	@go build ./...
	@echo "Build Done!"

.PHONY: deploy
deploy:
	@echo "Deploy CIPHER Fabric Chaincode"
	@cd ../../fabric-samples/test-network
	@echo "pwd: `pwd`"
	@./network.sh deployCC -ccn abac -ccp ../../chaincode/abac -ccl go