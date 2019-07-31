null  :=
space := $(null) #
comma := ,

EXLIST = $(foreach i,binance huobi okex,exchange/apifactory/$(i)/api exchange/apifactory/$(i)/internal)
MAINLIST = $(foreach i,ws message channle apifactory,exchange/$(i))
PKGSLIST = exchange $(MAINLIST) $(EXLIST)
COVERPKGS= $(subst $(space),$(comma),$(strip $(foreach i,$(PKGSLIST),github.com/sudachen/coin-exchange/$(i))))

build:
	cd exchange; go build

run-tests:
	cd tests && go test -coverprofile=../c.out -coverpkg=$(COVERPKGS)
