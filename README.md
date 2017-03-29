# Description
This is a command line and scriptable interface to the TI CC2538/CC26xx Serial Bootloader.

# Example Full Programming and Reset
1. Plug in a CC2538/CC26xx device with bootloader triggered.
   Say the serial bridge enumerates as _/dev/ttyUSB0_.
2. Open a terminal in a Code Composer Studio (CCS) project's __Debug__ directory.
3. Run `ccbootutil -verbose /dev/ttyUSB0 prgm SOME_PROJECT_NAME.out` .

Note: CCS generates the .out ELF file after a successful build of the Debug target.

# Status
Working!
