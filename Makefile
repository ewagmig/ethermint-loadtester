.PHONY: build install

# Determine the operating system
UNAME_S := $(shell uname -s)

# Define the date command based on the operating system
ifeq ($(UNAME_S),Darwin)  # MacOS
    DATE_CMD := gdate
else                      # Other Unix/Linux
    DATE_CMD := date
endif

# Define a logging function using the appropriate date command
define log
	@echo "[$$($(DATE_CMD) '+%m-%d %H:%M:%S.%3N')] $1"
endef

build:
	$(call log,start go build)
	@go build
	$(call log,done go build)

install: build
	$(call log,start go install)
	@go install
	$(call log,done go install)
