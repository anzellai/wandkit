# WandKit POC

This is an unofficial library interacting with Kano Wand Kit.

## Summary

I created this small library to interact with Kano's Wand Kit's BLE api.
I wanted to make the Wand as an HID like a mouse, good for presentation purposes -- click or flick the Wand and the presentation goes along!

I'm using a very nice fork from [Go ble](https://github.com/go-ble/ble) for the Bluetooth LE interfacing.

## Installation

`go get -u github.com/anzellai/wandkit`

If you have Go installed, simply run `make` should build MacOS binary to `/bin` folder.
Darwin MacOS is prebuilt on `/bin` directory.

I haven't mapped to *Linux* yet, and *Windows* requires a lot of work for the `Device` interface, perhaps a better approach is to use QT Bluetooth module as it's more cross platform compatible ASAIK.

Apart from the cmd binary, `wandkit` is also an independent library available to import as `"github.com/anzellai/wandkit/wandkit"`.


### Running

Simply run the compiled binary `./bin/wandkit-darwin` should work, it will create new device interface, connect to the first Kano Wand, and start exploring service/characteristics -- I have only selected `UserButton` and `SensorQuaternions` characteristics for simulating computer mouse HID.

More work to do on the calculation with quaternions to mouse XY 2d movement. I'm also thinking to only use gesture to simulate keyboard Arrow Keys mapping as it will be more reliable.


Anyhow, let's start poking fun with our Kano Wand Kit, and happy hacking!
