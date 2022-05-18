PROJECT_DIR=$(shell pwd)

define build_app
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(PROJECT_DIR)/builds/$(1) $(PROJECT_DIR)/cmd/$(1)/main.go
endef

clean:
	find $(shell pwd) \( -regex '.*mock_.*' -and ! -path "*/vendor/*" \) -exec rm {} \;
	find $(shell pwd) \( -path "*/mocks" -and ! -path "*/vendor/*" \) -exec rm -rf {} +
	rm -rf $(PROJECT_DIR)/vendor

vendor:
	GOPROXY=direct GOSUMDB=off go mod vendor

test:
	$(GOTEST) -v ./... -failfast

start:
	docker-compose up -d

up:
	docker-compose up -d --build --remove-orphans

down:
	docker-compose down

restart: down up

restart-service:
	docker-compose rm -fsv $(S)
	docker image prune -f
	$(call build_app,$(shell echo $(S) | sed "s/\_/\-/g"))
	docker-compose up -d --force-recreate --remove-orphans --build $(S)

logs:
	docker-compose logs -f $(S)

exec:
	docker-compose exec $(S) sh

ps:
	docker-compose ps

generate:
	go generate ./...

clean-report-dir:
	@rm -f $(REPORT_PATH)
	@mkdir -p $(REPORT_DIR)
	@touch $(COVERAGE_PATH_RAW)
	@touch $(REPORT_PATH)

run-test-with-coverage:
	$(GOTEST) -v $(shell go list ./... | grep -v "mocks" | grep -v "common") -v -coverprofile=$(COVERAGE_PATH_RAW) > $(REPORT_PATH).out

convert-report:
	cat $(COVERAGE_PATH_RAW) | grep -v "mocks" | grep -v "common" | grep -v "async_tasks_listener" > $(COVERAGE_PATH_RAW).tpm
	cat $(REPORT_PATH).out | $(GOBIN)/go-junit-report > $(REPORT_PATH)
	$(GOBIN)/gocov convert $(COVERAGE_PATH_RAW).tpm | $(GOBIN)/gocov-xml > $(COVERAGE_PATH)

process-report:
	$(foreach COVERAGE_PACKAGE,$(shell cat $(COVERAGE_PATH) | $(XML) sel -T -t -v "coverage/packages/package/@name"),$(call process_coverage,$(COVERAGE_PACKAGE),$(shell $(XML) sel -T -t -v "coverage/packages/package[@name='$(COVERAGE_PACKAGE)']/@line-rate" $(COVERAGE_PATH))))
	$(foreach TESTSUITE_PACKAGE,$(shell cat $(REPORT_PATH) | $(XML) sel -T -t -v "testsuites/testsuite/@name"), $(call process_testsuite,$(TESTSUITE_PACKAGE),$(shell echo $(TESTSUITE_PACKAGE) | sed "s/\//./g"));)

update: clean generate vendor

generate_unit: generate vendor

run-calc:
	go run $(PROJECT_DIR)/cmd/calc/main.go -c $(PROJECT_DIR)/common/config.yml

# ==============================================================================
# Administration
migrate:
	go run cmd/main.go admin migrate-up -c $(PROJECT_DIR)/common/config.yml

genkeys:
	go run cmd/main.go admin genkeys -c $(PROJECT_DIR)/common/config.yml

