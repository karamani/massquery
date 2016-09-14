configure:
	curl https://glide.sh/get | sh
	glide install
	cd /tmp && rm -f -R semverbuild && git clone https://github.com/karamani/semverbuild && cp semverbuild/svbuild $$GOPATH/bin/svbuild

update:
	glide up

compile:
	goimports -w ./*.go
	go tool vet ./*.go
	golint
	go test
	go install

build:
	$(eval newver := $(shell (svbuild -$(VER_LVL) | tail -n 1)))
	@echo $(newver)

	if [ -z "$(newver)" ] ; then \
		echo "Не могу создать версию"; \
	else \
		go build -ldflags "-X main.Version=$(newver)" -o "bin/massquery"; \
	fi

patch: export VER_LVL = l3
patch: compile build

minor: export VER_LVL = l2
minor: compile build

major: export VER_LVL = l1
major: compile build
