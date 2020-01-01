
all: app
dev: all
	spk dev
clean:
	rm -f app *.spk

app: $(shell find * -type f -name '*.go')
	go build -o $@
app.spk: app sandstorm-pkgdef.capnp
	spk pack $@

.PHONY: all clean dev
