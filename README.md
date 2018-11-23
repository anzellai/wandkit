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

### Ubuntu's experiment
I tried to run this project on Linux (specifically Ubuntu)

First I installed Go with different tutorials as a references, here some of them:

* [medium install go on Ubunto](https://medium.com/@patdhlk/how-to-install-go-1-9-1-on-ubuntu-16-04-ee64c073cd79)
* [linode install go on Ubunto](https://www.linode.com/docs/development/go/install-go-on-ubuntu/)

Then I run this commands:
```
cd cmd
go get
```
and eventually I found this error:

```
# github.com/go-vgo/robotgo/vendor/github.com/robotn/gohook
In file included from ./event/pub.h:20:0,
                 from ./event/goEvent.h:16,
                 from ../../github.com/go-vgo/robotgo/vendor/github.com/robotn/gohook/hook.go:22:
./event/../hook/x11/input_c.h:19:10: fatal error: X11/keysym.h: No such file or directory
 #include <X11/keysym.h>
          ^~~~~~~~~~~~~~
compilation terminated.
# github.com/raff/goble/xpc
In file included from ../../github.com/raff/goble/xpc/xpc.go:4:0:
./xpc_wrapper.h:6:10: fatal error: xpc/xpc.h: No such file or directory
 #include <xpc/xpc.h>
          ^~~~~~~~~~~
compilation terminated.

```
My conclusions are that is not possibile, at the moment, to run this project on Linux.