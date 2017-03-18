#!/bin/bash

DEVICE=/dev/ttyUSB0
#DEVICE=/dev/ttyAMA0
#DEVICE=/dev/ttyACM0
#FLAGS=-verbose
FLAGS=

BIN=./ccbootutil
BUILD="go build"

CMD="$BIN $FLAGS $DEVICE"

alias ping="$BUILD && $CMD ping"
alias rst="$BUILD && $CMD reset"
alias syc="$BUILD && $CMD sync"
alias flash="$BUILD && $CMD flash"
alias verify="$BUILD && $CMD verify"
alias getchipid="$BUILD && $CMD getchipid"
alias getstatus="$BUILD && $CMD getstatus"
alias bankerase="$BUILD && $CMD bankerase"
alias download="$BUILD && $CMD download"
alias setccfg="$BUILD && $CMD setccfg"

BL_PIN=22
#BL_PIN=5
BL_LEVEL=0 # 0=active low, 1=active high

ccfgargs="ID_BL_BACKDOOR_EN 0xC5 ID_BL_BACKDOOR_PIN $BL_PIN ID_BL_BACKDOOR_LEVEL $BL_LEVEL ID_BL_ENABLE 0xC5"

# WARNING - This cannot fix the CCFG, since they are in FLASH and FLASH can only be changed from 1 to 0. So, this would only do an AND with the current value.
fixccfg() {
	echo setccfg $ccfgargs
	setccfg $ccfgargs
}
