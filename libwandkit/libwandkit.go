package libwandkit

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/darwin"
	"github.com/go-ble/ble/linux"
	"github.com/go-vgo/robotgo"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	// Wand device name: Kano-Wand-XX-XX-XX
	Wand = "Kano-Wand"
)

// INFO
const (
	// BleUUIDInformationService ...
	BleUUIDInformationService = "64a70010f6914b93a6f40968f5b648f8"
	// BleUUIDInformationOrganisationChar ...
	BleUUIDInformationOrganisationChar = "64a7000bf6914b93a6f40968f5b648f8"
	// BleUUIDInformationSwChar ...
	BleUUIDInformationSwChar = "64a70013f6914b93a6f40968f5b648f8"
	// BleUUIDInformationHwChar ...
	BleUUIDInformationHwChar = "64a70001f6914b93a6f40968f5b648f8"
)

// IO
const (
	// BleUUIDIOService ...
	BleUUIDIOService = "64a70012f6914b93a6f40968f5b648f8"
	// BleUUIDIOBatteryChar ...
	BleUUIDIOBatteryChar = "64a70007f6914b93a6f40968f5b648f8"
	// BleUUIDIOUserButtonChar ...
	BleUUIDIOUserButtonChar = "64a7000df6914b93a6f40968f5b648f8"
	// BleUUIDIOVibratorChar ...
	BleUUIDIOVibratorChar = "64a70008f6914b93a6f40968f5b648f8"
	// BleUUIDIOLedChar ...
	BleUUIDIOLedChar = "64a70009f6914b93a6f40968f5b648f8"
	// BleUUIDIOKeepAliveChar ...
	BleUUIDIOKeepAliveChar = "64a7000ff6914b93a6f40968f5b648f8"
)

// SENSOR
const (
	// BleUUIDSensorService ...
	BleUUIDSensorService = "64a70011f6914b93a6f40968f5b648f8"
	// BleUUIDSensorQuaternionsChar ...
	BleUUIDSensorQuaternionsChar = "64a70002f6914b93a6f40968f5b648f8"
	// BleUUIDSensorRawChar ...
	BleUUIDSensorRawChar = "64a7000af6914b93a6f40968f5b648f8"
	// BleUUIDSensorMotionChar ...
	BleUUIDSensorMotionChar = "64a7000cf6914b93a6f40968f5b648f8"
	// BleUUIDSensorMagnCalibrateChar ...
	BleUUIDSensorMagnCalibrateChar = "64a70021f6914b93a6f40968f5b648f8"
	// BleUUIDSensorQuaternionsResetChar ...
	BleUUIDSensorQuaternionsResetChar = "64a70004f6914b93a6f40968f5b648f8"
	// BleUUIDSensorTempChar ...
	BleUUIDSensorTempChar = "64a70014f6914b93a6f40968f5b648f8"
)

type quaternion []uint16

// WandKit struct...
type WandKit struct {
	device        ble.Device
	logger        *log.Entry
	duration      time.Duration
	cln           ble.Client
	p             *ble.Profile
	button        bool
	actions       []string
	motions       []quaternion
	origins       [3][3]uint16
	subscriptions []*ble.Characteristic
}

// New return new instance of WK
func New(l *log.Entry, d time.Duration) *WandKit {
	device, err := NewDevice()
	if err != nil {
		l.Fatalf("can't create new device: %v", err)
	}
	wk := &WandKit{
		device:        device,
		logger:        l,
		duration:      d,
		subscriptions: []*ble.Characteristic{},
	}
	ble.SetDefaultDevice(wk.device)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT)
	go func() {
		<-sigc
		if wk.p != nil && wk.subscriptions != nil && len(wk.subscriptions) > 0 {
			for _, subscription := range wk.subscriptions {
				subLogger := wk.logger.WithFields(log.Fields{
					"subscription":        subscription.UUID.String(),
					"subscription_name":   ble.Name(subscription.UUID),
					"subscription_handle": fmt.Sprintf("0x%02X", subscription.Handle),
				})
				if err := wk.cln.Unsubscribe(subscription, false); err != nil {
					subLogger.Fatalf("unsubscribe error: %v", err)
				}
				subLogger.Info("subscription unsubscribed")
			}
		}
		defer func() {
			wk.cln = nil
			wk.p = nil
		}()
		err := wk.cln.CancelConnection()
		if err != nil {
			wk.logger.Errorf("can't disconnect: %v", err)
			os.Exit(-1)
		}
		wk.logger.Info("disconnected")
		os.Exit(0)
	}()
	return wk
}

// NewDevice return new Ble Device instance
func NewDevice() (d ble.Device, err error) {
	switch runtime.GOOS {
	case "linux":
		return DefaultLinuxDevice()
	case "windows":
		return nil, errors.New("not implemented")
	default:
		return DefaultDarwinDevice()
	}
}

// DefaultLinuxDevice interface...
func DefaultLinuxDevice() (d ble.Device, err error) {
	return linux.NewDevice()
}

// DefaultDarwinDevice interface...
func DefaultDarwinDevice() (d ble.Device, err error) {
	return darwin.NewDevice()
}

// Connect will scan and get WandKit device
func (wk *WandKit) Connect() {
	filter := func(a ble.Advertisement) bool {
		return strings.HasPrefix(strings.ToUpper(a.LocalName()), strings.ToUpper(Wand))
	}
	wk.logger.Infof("scanning %s for %s", Wand, wk.duration)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), wk.duration))
	cln, err := ble.Connect(ctx, filter)
	if err != nil {
		wk.logger.Fatalf("can't connect to %s: %v", Wand, err)
	}
	wk.cln = cln

	wk.logger.Info("discovering profile")
	p, err := cln.DiscoverProfile(true)
	if err != nil {
		wk.logger.Fatalf("can't discover profile: %v", err)
	}
	wk.logger.Infof("profile discovered: %+v", p)
	wk.p = p
}

// Explore will explore BLE instance with subscriptions
func (wk *WandKit) Explore() {
	wk.logger.Info("connector start")
	for _, s := range wk.p.Services {
		srvLogger := wk.logger.WithFields(log.Fields{
			"service":        s.UUID.String(),
			"service_name":   ble.Name(s.UUID),
			"service_handle": fmt.Sprintf("0x%02X", s.Handle),
		})
		srvLogger.Info("service discovered")

		for _, c := range s.Characteristics {
			chrLogger := srvLogger.WithFields(log.Fields{
				"characteristic":           c.UUID.String(),
				"characteristic_name":      ble.Name(c.UUID),
				"characteristic_property":  propString(c.Property),
				"characteristics_handle":   fmt.Sprintf("0x%02X", c.Handle),
				"characteristics_v_handle": fmt.Sprintf("0x%02X", c.ValueHandle),
			})
			chrLogger.Info("characteristic discovered")

			charUUID := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%s", c.UUID)))
			// We are only interested in User Button & Sensor Quaternions characteristics
			if !(charUUID == BleUUIDIOUserButtonChar || charUUID == BleUUIDSensorQuaternionsChar) {
				continue
			}

			// Don't bother to subscribe the Service Changed characteristics.
			if c.UUID.Equal(ble.ServiceChangedUUID) {
				continue
			}
			// Don't touch the Apple-specific Service/Characteristic.
			// Service: D0611E78BBB44591A5F8487910AE4366
			// Characteristic: 8667556C9A374C9184ED54EE27D90049, Property: 0x18 (WN),
			//   Descriptor: 2902, Client Characteristic Configuration
			//   Value         0000 | "\x00\x00"
			if c.UUID.Equal(ble.MustParse("8667556C9A374C9184ED54EE27D90049")) {
				continue
			}

			if c.Property&ble.CharNotify != 0 {
				chrLogger.Infof("subscribe to notification for %s", wk.duration)
				if err := wk.cln.Subscribe(c, false, onCallback(wk)); err != nil {
					chrLogger.Fatalf("subscribe error: %v", err)
				}
				wk.subscriptions = append(wk.subscriptions, c)
			}

			for _, d := range c.Descriptors {
				dspLogger := chrLogger.WithFields(log.Fields{
					"descriptor":        d.UUID.String(),
					"descriptor_name":   ble.Name(d.UUID),
					"descriptor_handle": fmt.Sprintf("0x%02X", d.Handle),
				})
				dspLogger.Info("descriptor discovered")

				b, err := wk.cln.ReadDescriptor(d)
				if err != nil {
					dspLogger.Errorf("read error: %v", err)
					continue
				}
				dspLogger.Infof("value read: %x | %q", b, b)
			}
		}
	}
}

// Motion calculates if a keyboard arrow should trigger
func (wk *WandKit) Motion(w, x, y, z uint16) (action string) {
	action = "noop"
	var avgW, avgX, avgY, avgZ uint16
	vectorDir := [3]uint16{}
	vectorUp := [3]uint16{}
	vectorLeft := [3]uint16{}
	if len(wk.motions) < 10 {
		return
	}
	for _, motion := range wk.motions {
		avgW += motion[0]
		avgX += motion[1]
		avgY += motion[2]
		avgZ += motion[3]
	}
	avgW = avgW / uint16(len(wk.motions))
	avgX = avgX / uint16(len(wk.motions))
	avgY = avgY / uint16(len(wk.motions))
	avgZ = avgZ / uint16(len(wk.motions))

	vectorDir[0] = 2 * (avgX*avgZ + avgW*avgY)
	vectorDir[1] = 2 * (avgY*avgZ - avgW*avgW)
	vectorDir[2] = 1 - 2*(avgX*avgX+avgY*avgY)
	vectorUp[0] = 2 * (avgX*avgY - avgW*avgZ)
	vectorUp[1] = 1 - 2*(avgX*avgX+avgZ*avgZ)
	vectorUp[2] = w * (avgY*avgZ + avgW*avgX)
	vectorLeft[0] = 1 - 2*(avgY*avgY+avgZ*avgZ)
	vectorLeft[1] = 2 * (avgX*avgY + avgW*avgZ)
	vectorLeft[2] = 2 * (avgX*avgZ + avgW*avgY)
	if wk.origins[0][0] == 0 && wk.origins[0][1] == 0 && wk.origins[0][2] == 0 {
		wk.origins[0] = vectorDir
		wk.origins[1] = vectorUp
		wk.origins[2] = vectorLeft
	}
	if wk.origins[1][0] > vectorUp[0] && wk.origins[1][1] > vectorUp[1] && wk.origins[1][2] > vectorUp[2] {
		action = "up"
	} else if wk.origins[1][0] < vectorUp[0] && wk.origins[1][1] < vectorUp[1] && wk.origins[1][2] < vectorUp[2] {
		action = "down"
	} else if wk.origins[2][0] > vectorLeft[0] && wk.origins[2][1] > vectorLeft[1] && wk.origins[2][2] > vectorLeft[2] {
		action = "Left"
	} else if wk.origins[2][0] < vectorLeft[0] && wk.origins[2][1] < vectorLeft[1] && wk.origins[2][2] < vectorLeft[2] {
		action = "Right"
	}
	wk.logger.WithFields(log.Fields{
		"position": []uint16{w, x, y, z},
		"average":  []uint16{avgW, avgX, avgY, avgZ},
		"origins":  wk.origins,
		"vectors":  [][3]uint16{vectorDir, vectorUp, vectorLeft},
	}).Debugf("motion: [%s]", action)
	if action != "noop" {
		// reset the origins and accept next action
		wk.origins[0][0], wk.origins[0][1], wk.origins[0][2] = 0, 0, 0
		keyboardArrow(wk.logger, action)
	}
	return
}

func onCallback(wk *WandKit) func([]byte) {
	cbLogger := wk.logger.WithFields(log.Fields{
		"callback": "onData",
	})
	return func(req []byte) {
		if wk.p == nil || wk.subscriptions == nil || len(wk.subscriptions) == 0 {
			return
		}
		// UserButton
		if len(req) <= 2 {
			wk.button = req[0] == 1
			if wk.button {
				mouseClick(cbLogger, req[0])
			} else {
				wk.motions = nil
			}
			cbLogger.Debugf("user-button: [%d]", req[0])
		} else
		// Sensor Quaternions
		if len(req) == 8 && wk.button {
			w, x, y, z := ToUint16(req[0], req[1]), ToUint16(req[2], req[3]), ToUint16(req[4], req[5]), ToUint16(req[6], req[7])
			wk.motions = append(wk.motions, []uint16{w, x, y, z})
			if len(wk.motions) > 10 {
				wk.motions = wk.motions[1:]
			}
			cbLogger.Debugf("position: [%d, %d, %d, %d]", w, x, y, z)
			motion := wk.Motion(w, x, y, z)
			if motion != "noop" {
				wk.actions = append(wk.actions, motion)
				if len(wk.actions) > 2 {
					for _, action := range wk.actions {
						if action != motion {
							wk.actions = nil
							wk.motions = nil
							return
						}
					}
					// FIXME: Disable for now until useful!
					keyboardArrow(cbLogger, motion)
					wk.motions = nil
				}
			}
		}
	}
}

func mouseClick(logger *log.Entry, input byte) {
	if input > 0 {
		robotgo.MouseClick("left", true)
		logger.Debugf("mouse-click: [%d]", input)
	}
}

func keyboardArrow(logger *log.Entry, input string) {
	robotgo.KeyTap(input)
	logger.Debugf("keyboard-tap: [%s]", input)
}

// ToUint16 helper function to convert 2 bytes to Uint16
func ToUint16(a, b byte) uint16 {
	return uint16(a)<<8 | uint16(b)
}

func propString(p ble.Property) string {
	var s string
	for k, v := range map[ble.Property]string{
		ble.CharBroadcast:   "B",
		ble.CharRead:        "R",
		ble.CharWriteNR:     "w",
		ble.CharWrite:       "W",
		ble.CharNotify:      "N",
		ble.CharIndicate:    "I",
		ble.CharSignedWrite: "S",
		ble.CharExtended:    "E",
	} {
		if p&k != 0 {
			s += v
		}
	}
	return s
}
