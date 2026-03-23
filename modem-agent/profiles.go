package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xlab/at"
)

// GenericProfile works with any modem that supports standard 3GPP AT commands.
// It skips Huawei-specific init (AT^SYSINFO, AT+COPS format, AT+GMM, AT+GSN)
// and optionally unlocks the SIM with a PIN.
type GenericProfile struct {
	at.DefaultProfile
	simPIN string
}

func (p *GenericProfile) Init(d *at.Device) error {
	p.Dev = d
	d.State = &at.DeviceState{}

	// Flush
	d.Send(at.NoopCmd)

	// Unlock SIM if PIN is provided
	if p.simPIN != "" {
		if _, err := d.Send("AT+CPIN=" + p.simPIN); err != nil {
			return fmt.Errorf("SIM PIN unlock failed: %w", err)
		}
		time.Sleep(2 * time.Second)
	}

	// Standard SMS init sequence (no vendor-specific commands)
	if err := p.CMGF(false); err != nil {
		return fmt.Errorf("at init: unable to switch message format to PDU: %w", err)
	}
	if err := p.CPMS(at.MemoryTypes.NvRAM, at.MemoryTypes.NvRAM, at.MemoryTypes.NvRAM); err != nil {
		return fmt.Errorf("at init: unable to set messages storage: %w", err)
	}
	if err := p.CNMI(1, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("at init: unable to turn on message notifications: %w", err)
	}
	if err := p.CLIP(true); err != nil {
		return fmt.Errorf("at init: unable to turn on calling party ID notifications: %w", err)
	}

	return p.FetchInbox()
}

// Huawei-specific no-ops — these are called by handleReport (via Watch)
// but have no meaning on generic modems.
func (p *GenericProfile) SYSINFO() (*at.SystemInfoReport, error) { return nil, nil }
func (p *GenericProfile) BOOT(token uint64) error               { return nil }
func (p *GenericProfile) SYSCFG(roaming, cellular bool) error    { return nil }
func (p *GenericProfile) COPS(auto bool, text bool) error        { return nil }

// HuaweiProfile extends DefaultProfile with SIM PIN support.
// After optional PIN unlock it delegates to the full Huawei init (AT^SYSINFO, etc.).
type HuaweiProfile struct {
	at.DefaultProfile
	simPIN string
}

func (p *HuaweiProfile) Init(d *at.Device) error {
	if p.simPIN != "" {
		p.Dev = d
		d.Send(at.NoopCmd) // flush
		if _, err := d.Send("AT+CPIN=" + p.simPIN); err != nil {
			return fmt.Errorf("SIM PIN unlock failed: %w", err)
		}
		time.Sleep(2 * time.Second)
	}
	return p.DefaultProfile.Init(d)
}

// probeProfile is a minimal profile used only for modem detection.
// It initialises just enough for Send() to work without running any AT init sequence.
type probeProfile struct {
	at.DefaultProfile
}

func (p *probeProfile) Init(d *at.Device) error {
	p.Dev = d
	d.State = &at.DeviceState{}
	return nil
}

// resolveProfile returns the appropriate DeviceProfile for the given name.
func resolveProfile(name, simPIN string) at.DeviceProfile {
	switch name {
	case "huawei-e173":
		return &HuaweiProfile{simPIN: simPIN}
	default:
		return &GenericProfile{simPIN: simPIN}
	}
}

// detectProfile probes the modem with ATI and returns a profile name
// based on the manufacturer/model response.
func detectProfile(dev *at.Device) string {
	if err := dev.Init(&probeProfile{}); err != nil {
		log.Printf("probe init (non-fatal): %v", err)
	}

	reply, err := dev.Send("ATI")
	if err != nil {
		return "generic"
	}

	upper := strings.ToUpper(reply)
	if strings.Contains(upper, "HUAWEI") || strings.Contains(upper, "E173") {
		return "huawei-e173"
	}
	return "generic"
}
