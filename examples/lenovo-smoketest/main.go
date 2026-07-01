// Command lenovo-smoketest runs an opt-in, real-hardware smoke test of the
// Lenovo XCC bmclib provider, and (in -vmedia-boot mode) the ordered
// virtual-media boot flow.
//
// By default it is READ-ONLY: it connects, runs the vendor compatibility check,
// and exercises every read capability (power/boot/BIOS/secure-boot/thermal/
// power/inventory/storage/users/logs/license/SKLM/network/serial/events/
// telemetry/jobs/certificates/SNMP/firmware-steps), printing a PASS/WARN/FAIL/
// SKIP report.
//
// Mutating operations are NEVER performed unless the master -allow-writes flag
// is set AND the specific operation's flag is provided. This makes accidental
// changes to production hardware impossible with a default invocation.
//
// Read failures on capabilities a given XCC model may not expose (telemetry,
// jobs, SKLM, ...) are reported as WARN (non-fatal). Failures of the core
// capabilities (compatibility, service root, power state, inventory) are FAIL
// and set a non-zero exit code.
//
// Usage:
//
//	# read-only sweep
//	go run ./examples/lenovo-smoketest -host <ip> -user <user> -pass <pass>
//
//	# include selected mutations (each still requires its own flag)
//	go run ./examples/lenovo-smoketest -host <ip> -user <user> -pass <pass> \
//	    -allow-writes -power on -bootdev pxe
//
//	# virtual-media boot: mount an ISO, set it as next boot device, power on
//	# (one ordered flow; requires -allow-writes; skips the read/write sweep)
//	go run ./examples/lenovo-smoketest -host <ip> -user <user> -pass <pass> \
//	    -allow-writes -vmedia-boot -vmedia-url http://10.0.0.2/installer.iso \
//	    -vmedia-kind CD -bootdev-efi -power on
//
// This is intended for manual validation against real hardware; it is not part
// of the unit-test suite (which runs entirely against fixtures). See RUNBOOK.md
// and HARDWARE-TEST-PLAN.md.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	"github.com/bmc-toolbox/bmclib/v2/providers/lenovo"
	logrusr "github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"
)

// ---------------------------------------------------------------------------
// harness
// ---------------------------------------------------------------------------

const (
	statusPass = "PASS"
	statusWarn = "WARN"
	statusFail = "FAIL"
	statusSkip = "SKIP"
)

// result is a single smoke-test check outcome.
type result struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// harness drives the checks against a connected XCC and accumulates results.
type harness struct {
	conn    *lenovo.Conn
	ctx     context.Context
	verbose bool
	results []result
}

// record appends a result and prints it as it happens.
func (h *harness) record(name, status, detail string) {
	h.results = append(h.results, result{Name: name, Status: status, Detail: detail})
	fmt.Printf("[%s] %-34s %s\n", status, name, detail)
}

// read runs a read-only check. On error the status is FAIL when critical (core
// capability), otherwise WARN (the XCC model may not expose the resource).
func (h *harness) read(name string, critical bool, fn func() (string, error)) {
	summary, err := fn()
	if err != nil {
		status := statusWarn
		if critical {
			status = statusFail
		}
		h.record(name, status, err.Error())
		return
	}
	h.record(name, statusPass, summary)
}

// action runs a mutating check. It is skipped unless enabled is true (the
// operation's flag was given together with -allow-writes).
func (h *harness) action(name string, enabled bool, skipReason string, fn func() error) {
	if !enabled {
		h.record(name, statusSkip, skipReason)
		return
	}
	if err := fn(); err != nil {
		h.record(name, statusFail, err.Error())
		return
	}
	h.record(name, statusPass, "ok")
}

// ---------------------------------------------------------------------------
// flags
// ---------------------------------------------------------------------------

// options holds the parsed command-line configuration.
type options struct {
	host, user, pass, port string
	basic                  bool
	verbose                bool
	jsonOut                bool
	timeout                time.Duration

	// write gate + per-operation flags (all empty/false => skipped).
	allowWrites    bool
	power          string
	bootDevice     string
	bootPersistent bool
	bootEFI        bool
	nmi            bool
	bmcReset       string
	setBIOS        string
	resetBIOS      bool
	secureBoot     string
	powerCap       string
	vmediaURL      string
	vmediaKind     string
	vmediaEject    bool
	vmediaBoot     bool
	clearSEL       bool
	snmpV1Trap     string
	snmpV3Trap     string

	firmwareFile      string
	firmwareComponent string
	firmwareApplyTime string
	firmwarePollEvery time.Duration
}

func parseFlags() options {
	var o options
	flag.StringVar(&o.host, "host", "", "XCC hostname/IP (required)")
	flag.StringVar(&o.user, "user", "", "username (required)")
	flag.StringVar(&o.pass, "pass", "", "password (required)")
	flag.StringVar(&o.port, "port", "443", "XCC Redfish port")
	flag.BoolVar(&o.basic, "basic", false, "use HTTP Basic auth instead of a session")
	flag.BoolVar(&o.verbose, "v", false, "verbose: dump full payloads for read checks")
	flag.BoolVar(&o.jsonOut, "json", false, "emit the result report as JSON")
	flag.DurationVar(&o.timeout, "timeout", 5*time.Minute, "overall timeout")

	flag.BoolVar(&o.allowWrites, "allow-writes", false, "MASTER SWITCH: permit mutating operations (each still needs its own flag)")
	flag.StringVar(&o.power, "power", "", "set power state: on|off|soft|cycle|reset")
	flag.StringVar(&o.bootDevice, "bootdev", "", "set boot device: pxe|disk|cdrom|bios|usb|floppy|remote_drive|... (CD/DVD/USBStick aliases accepted)")
	flag.BoolVar(&o.bootPersistent, "bootdev-persistent", false, "make -bootdev / -vmedia-boot persistent (Continuous)")
	flag.BoolVar(&o.bootEFI, "bootdev-efi", false, "make -bootdev / -vmedia-boot UEFI mode")
	flag.BoolVar(&o.nmi, "nmi", false, "send an NMI to the host")
	flag.StringVar(&o.bmcReset, "bmc-reset", "", "reset the BMC: GracefulRestart|ForceRestart")
	flag.StringVar(&o.setBIOS, "set-bios", "", "set BIOS attributes: \"Key=Value,Key2=Value2\"")
	flag.BoolVar(&o.resetBIOS, "reset-bios", false, "reset BIOS settings to defaults")
	flag.StringVar(&o.secureBoot, "secureboot", "", "set Secure Boot: enable|disable")
	flag.StringVar(&o.powerCap, "powercap", "", "set power cap watts, or \"off\" to disable")
	flag.StringVar(&o.vmediaURL, "vmedia-url", "", "insert virtual media from this URL")
	flag.StringVar(&o.vmediaKind, "vmedia-kind", "CD", "virtual media kind for -vmedia-url / -vmedia-boot: CD|DVD|Floppy|USBStick")
	flag.BoolVar(&o.vmediaEject, "vmedia-eject", false, "eject virtual media of -vmedia-kind")
	flag.BoolVar(&o.vmediaBoot, "vmedia-boot", false, "ordered flow: mount -vmedia-url, set it as next boot device, power on (skips the read/write sweep)")
	flag.BoolVar(&o.clearSEL, "clear-sel", false, "clear the System Event Log")
	flag.StringVar(&o.snmpV1Trap, "snmp-v1-trap", "", "enable|disable the SNMPv1 trap")
	flag.StringVar(&o.snmpV3Trap, "snmp-v3-trap", "", "enable|disable the SNMPv3 trap")
	flag.StringVar(&o.firmwareFile, "firmware", "", "firmware image file to install (XCC push protocol)")
	flag.StringVar(&o.firmwareComponent, "firmware-component", "", "firmware target (FirmwareInventory id, e.g. BMC-Backup); empty = auto-detect")
	flag.StringVar(&o.firmwareApplyTime, "firmware-applytime", "OnReset", "firmware OperationApplyTime: Immediate|OnReset|OnStartUpdateRequest")
	flag.DurationVar(&o.firmwarePollEvery, "firmware-poll", 15*time.Second, "firmware task poll interval")
	flag.Parse()

	return o
}

func main() {
	o := parseFlags()
	if o.host == "" || o.user == "" || o.pass == "" {
		fmt.Fprintln(os.Stderr, "error: -host, -user and -pass are required")
		flag.Usage()
		os.Exit(2)
	}

	l := logrus.New()
	if o.verbose {
		l.Level = logrus.DebugLevel
	} else {
		l.Level = logrus.InfoLevel
	}
	logger := logrusr.New(l)

	ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
	defer cancel()

	opts := []lenovo.Option{lenovo.WithPort(o.port)}
	if o.basic {
		opts = append(opts, lenovo.WithUseBasicAuth(true))
	}
	conn := lenovo.New(o.host, o.user, o.pass, logger, opts...)

	h := &harness{conn: conn, ctx: ctx, verbose: o.verbose}

	// Compatibility opens and closes its own session.
	if !conn.Compatible(ctx) {
		h.record("connection.compatible", statusFail, "device is NOT compatible with the lenovo provider")
		h.finish(o)
		return
	}
	h.record("connection.compatible", statusPass, "device identifies as Lenovo XCC")

	if err := conn.Open(ctx); err != nil {
		h.record("connection.open", statusFail, err.Error())
		h.finish(o)
		return
	}
	h.record("connection.open", statusPass, "session established")
	defer conn.Close(ctx)

	// -vmedia-boot is a focused, ordered mutation flow; it replaces the
	// read/write sweep rather than running alongside it.
	if o.vmediaBoot {
		runVMediaBoot(h, o)
		h.finish(o)
		return
	}

	runReadChecks(h)
	runWriteChecks(h, o)

	h.finish(o)
}

// finish prints the report (text or JSON) and exits with the right code.
func (h *harness) finish(o options) {
	if o.jsonOut {
		_ = json.NewEncoder(os.Stdout).Encode(h.results)
	}

	var pass, warn, fail, skip int
	for _, r := range h.results {
		switch r.Status {
		case statusPass:
			pass++
		case statusWarn:
			warn++
		case statusFail:
			fail++
		case statusSkip:
			skip++
		}
	}

	fmt.Printf("\nsummary: %d PASS, %d WARN, %d FAIL, %d SKIP (%d checks)\n",
		pass, warn, fail, skip, len(h.results))

	if fail > 0 {
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// read checks
// ---------------------------------------------------------------------------

// runReadChecks exercises every read capability of the provider. Core
// capabilities are marked critical (failure => non-zero exit); the rest are
// non-critical (a given XCC model may not expose them, reported as WARN).
func runReadChecks(h *harness) {
	ctx := h.ctx
	c := h.conn

	h.read("serviceroot.read", true, func() (string, error) {
		root, err := c.ServiceRoot(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("vendor=%q product=%q redfish=%s", root.Vendor, root.Product, root.RedfishVersion), nil
	})

	// --- power / boot ---
	h.read("power.state", true, func() (string, error) {
		state, err := c.PowerStateGet(ctx)
		return "state=" + state, err
	})
	h.read("boot.progress", false, func() (string, error) {
		bp, err := c.GetBootProgress()
		if err != nil {
			return "", err
		}
		return "lastState=" + string(bp.LastState), nil
	})
	h.read("boot.override", false, func() (string, error) {
		o, err := c.BootDeviceOverrideGet(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("device=%s persistent=%t efi=%t", o.Device, o.IsPersistent, o.IsEFIBoot), nil
	})

	// --- bios / secure boot ---
	h.read("bios.read", false, func() (string, error) {
		cfg, err := c.GetBiosConfiguration(ctx)
		if err != nil {
			return "", err
		}
		// In verbose mode dump every attribute name=value so the operator can
		// copy exact keys for -set-bios (attribute names are model/firmware
		// specific; an unknown key fails the whole PATCH with PropertyUnknown).
		if h.verbose {
			keys := make([]string, 0, len(cfg))
			for k := range cfg {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("       bios attr %s = %s\n", k, cfg[k])
			}
		}
		return fmt.Sprintf("%d attributes", len(cfg)), nil
	})
	h.read("secureboot.read", false, func() (string, error) {
		sb, err := c.GetSecureBoot(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("enabled=%t mode=%s current=%s", sb.Enabled, sb.Mode, sb.CurrentBoot), nil
	})

	// --- thermal / power metrics ---
	h.read("thermal.read", false, func() (string, error) {
		t, err := c.Thermal(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d temps, %d fans", len(t.Temperatures), len(t.Fans)), nil
	})
	h.read("power.metrics", false, func() (string, error) {
		p, err := c.ReadPower(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("consumed=%.0fW capacity=%.0fW supplies=%d", p.ConsumedWatts, p.CapacityWatts, len(p.PowerSupplies)), nil
	})

	// --- inventory / storage ---
	h.read("inventory.read", true, func() (string, error) {
		dev, err := c.Inventory(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("vendor=%q model=%q cpus=%d dimms=%d drives=%d nics=%d",
			dev.Vendor, dev.Model, len(dev.CPUs), len(dev.Memory), len(dev.Drives), len(dev.NICs)), nil
	})

	var firstController string
	h.read("storage.controllers", false, func() (string, error) {
		ctrls, err := c.StorageControllers(ctx)
		if err != nil {
			return "", err
		}
		if len(ctrls) > 0 {
			firstController = ctrls[0].ID
		}
		return fmt.Sprintf("%d controllers", len(ctrls)), nil
	})
	if firstController != "" {
		h.read("storage.volumes", false, func() (string, error) {
			vols, err := c.Volumes(ctx, firstController)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("controller %s: %d volumes", firstController, len(vols)), nil
		})
	} else {
		h.record("storage.volumes", statusSkip, "no storage controller to query")
	}

	// --- users ---
	h.read("users.read", false, func() (string, error) {
		users, err := c.UserRead(ctx)
		return fmt.Sprintf("%d accounts", len(users)), err
	})

	// --- logs ---
	h.read("sel.read", false, func() (string, error) {
		entries, err := c.GetSystemEventLog(ctx)
		return fmt.Sprintf("%d SEL entries", len(entries)), err
	})
	h.read("log.audit", false, func() (string, error) {
		entries, err := c.EventLog(ctx, lenovo.LogServiceAudit)
		return fmt.Sprintf("%d audit entries", len(entries)), err
	})

	// --- bmc management ---
	h.read("license.read", false, func() (string, error) {
		lics, err := c.Licenses(ctx)
		return fmt.Sprintf("%d licenses", len(lics)), err
	})
	h.read("sklm.read", false, func() (string, error) {
		cfg, err := c.GetSecureKeyLifecycle(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("deviceGroup=%q keyServers=%d", cfg.DeviceGroup, len(cfg.KeyRepoServers)), nil
	})

	// --- network / serial ---
	h.read("network.bmc", false, func() (string, error) {
		ifaces, err := c.BMCNetworkInterfaces(ctx)
		return fmt.Sprintf("%d BMC NICs", len(ifaces)), err
	})
	h.read("network.server", false, func() (string, error) {
		ifaces, err := c.ServerNetworkInterfaces(ctx)
		return fmt.Sprintf("%d server NICs", len(ifaces)), err
	})
	h.read("network.protocols", false, func() (string, error) {
		ps, err := c.NetworkProtocols(ctx)
		return fmt.Sprintf("%d services", len(ps)), err
	})
	h.read("serial.read", false, func() (string, error) {
		ifaces, err := c.SerialInterfaces(ctx)
		return fmt.Sprintf("%d serial interfaces", len(ifaces)), err
	})

	// --- events / telemetry ---
	h.read("event.service", false, func() (string, error) {
		info, err := c.EventService(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("enabled=%t sse=%q", info.ServiceEnabled, info.SSEFilterURI), nil
	})
	h.read("event.subscriptions", false, func() (string, error) {
		subs, err := c.EventSubscriptions(ctx)
		return fmt.Sprintf("%d subscriptions", len(subs)), err
	})
	h.read("telemetry.service", false, func() (string, error) {
		info, err := c.TelemetryService(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("enabled=%t maxReports=%d", info.ServiceEnabled, info.MaxReports), nil
	})
	h.read("telemetry.reports", false, func() (string, error) {
		ids, err := c.MetricReports(ctx)
		return fmt.Sprintf("%d metric reports", len(ids)), err
	})
	h.read("telemetry.definitions", false, func() (string, error) {
		ids, err := c.MetricDefinitions(ctx)
		return fmt.Sprintf("%d metric definitions", len(ids)), err
	})

	// --- jobs ---
	h.read("job.service", false, func() (string, error) {
		info, err := c.JobService(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("enabled=%t", info.ServiceEnabled), nil
	})
	h.read("job.list", false, func() (string, error) {
		jobs, err := c.Jobs(ctx)
		return fmt.Sprintf("%d jobs", len(jobs)), err
	})

	// --- certificates ---
	var firstCert string
	h.read("cert.locations", false, func() (string, error) {
		locs, err := c.CertificateLocations(ctx)
		if err != nil {
			return "", err
		}
		if len(locs) > 0 {
			firstCert = locs[0]
		}
		return fmt.Sprintf("%d certificates", len(locs)), nil
	})
	if firstCert != "" {
		h.read("cert.read", false, func() (string, error) {
			cert, err := c.Certificate(ctx, firstCert)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("subject=%q notAfter=%s", cert.SubjectCommonName, cert.ValidNotAfter), nil
		})
	} else {
		h.record("cert.read", statusSkip, "no certificate locations to read")
	}

	// --- snmp ---
	h.read("snmp.read", false, func() (string, error) {
		cfg, err := c.SNMP(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("v1Trap=%t v3Trap=%t port=%d", cfg.V1TrapEnabled, cfg.V3TrapEnabled, cfg.TrapPort), nil
	})

	// --- firmware (read-only step plan) ---
	h.read("firmware.steps", false, func() (string, error) {
		steps, err := c.FirmwareInstallSteps(ctx, "bmc")
		return fmt.Sprintf("%d install steps", len(steps)), err
	})
}

// ---------------------------------------------------------------------------
// write checks
// ---------------------------------------------------------------------------

// gate returns whether a mutating operation may run. An operation runs only when
// its own flag is set AND the master -allow-writes switch is on.
func (o options) gate(flagSet bool) (bool, string) {
	switch {
	case !flagSet:
		return false, "flag not set"
	case !o.allowWrites:
		return false, "-allow-writes not set"
	default:
		return true, ""
	}
}

// runWriteChecks performs the gated mutating operations. With a default
// invocation (no -allow-writes) every check here is SKIPped.
func runWriteChecks(h *harness, o options) {
	ctx := h.ctx
	c := h.conn

	enabled, reason := o.gate(o.power != "")
	h.action("power.set", enabled, reason, func() error {
		ok, err := c.PowerSet(ctx, o.power)
		if err == nil && !ok {
			return fmt.Errorf("power set returned ok=false")
		}
		return err
	})

	enabled, reason = o.gate(o.bootDevice != "")
	h.action("boot.set", enabled, reason, func() error {
		dev, perr := normalizeBootDevice(o.bootDevice)
		if perr != nil {
			return perr
		}
		ok, err := c.BootDeviceSet(ctx, dev, o.bootPersistent, o.bootEFI)
		if err == nil && !ok {
			return fmt.Errorf("boot device set returned ok=false")
		}
		return err
	})

	enabled, reason = o.gate(o.nmi)
	h.action("power.nmi", enabled, reason, func() error {
		return c.SendNMI(ctx)
	})

	enabled, reason = o.gate(o.bmcReset != "")
	h.action("bmc.reset", enabled, reason, func() error {
		ok, err := c.BmcReset(ctx, o.bmcReset)
		if err == nil && !ok {
			return fmt.Errorf("bmc reset returned ok=false")
		}
		return err
	})

	enabled, reason = o.gate(o.setBIOS != "")
	h.action("bios.set", enabled, reason, func() error {
		attrs, perr := parseKV(o.setBIOS)
		if perr != nil {
			return perr
		}
		// Pre-validate attribute names against this system's BIOS registry. XCC
		// validates the PATCH atomically: a single unknown key fails the whole
		// request with an opaque PropertyUnknown/InternalError 500. Catch it
		// early with an actionable message (names are model/firmware-specific).
		if cur, cerr := c.GetBiosConfiguration(ctx); cerr == nil && len(cur) > 0 {
			var unknown []string
			for k := range attrs {
				if _, ok := cur[k]; !ok {
					unknown = append(unknown, k)
				}
			}
			if len(unknown) > 0 {
				sort.Strings(unknown)
				return fmt.Errorf("unknown BIOS attribute(s) %v on this system "+
					"(names are model-specific; run with -v and grep 'bios attr' for valid keys)", unknown)
			}
		}
		return c.SetBiosConfiguration(ctx, attrs)
	})

	enabled, reason = o.gate(o.resetBIOS)
	h.action("bios.reset", enabled, reason, func() error {
		return c.ResetBiosConfiguration(ctx)
	})

	enabled, reason = o.gate(o.secureBoot != "")
	h.action("secureboot.set", enabled, reason, func() error {
		on, perr := parseEnable(o.secureBoot)
		if perr != nil {
			return perr
		}
		return c.SetSecureBoot(ctx, on)
	})

	enabled, reason = o.gate(o.powerCap != "")
	h.action("power.cap", enabled, reason, func() error {
		limit, perr := parsePowerCap(o.powerCap)
		if perr != nil {
			return perr
		}
		return c.SetPowerCap(ctx, limit)
	})

	enabled, reason = o.gate(o.vmediaURL != "")
	h.action("vmedia.insert", enabled, reason, func() error {
		ok, err := c.SetVirtualMedia(ctx, o.vmediaKind, o.vmediaURL)
		if err == nil && !ok {
			return fmt.Errorf("set virtual media returned ok=false")
		}
		return err
	})

	enabled, reason = o.gate(o.vmediaEject)
	h.action("vmedia.eject", enabled, reason, func() error {
		ok, err := c.SetVirtualMedia(ctx, o.vmediaKind, "")
		if err == nil && !ok {
			return fmt.Errorf("eject virtual media returned ok=false")
		}
		return err
	})

	enabled, reason = o.gate(o.clearSEL)
	h.action("sel.clear", enabled, reason, func() error {
		return c.ClearSystemEventLog(ctx)
	})

	enabled, reason = o.gate(o.snmpV1Trap != "")
	h.action("snmp.v1trap", enabled, reason, func() error {
		on, perr := parseEnable(o.snmpV1Trap)
		if perr != nil {
			return perr
		}
		return c.EnableSNMPv1Trap(ctx, on)
	})

	enabled, reason = o.gate(o.snmpV3Trap != "")
	h.action("snmp.v3trap", enabled, reason, func() error {
		on, perr := parseEnable(o.snmpV3Trap)
		if perr != nil {
			return perr
		}
		return c.EnableSNMPv3Trap(ctx, on)
	})

	runFirmwareCheck(h, o)
}

// runFirmwareCheck performs the XCC firmware push (claim -> push -> poll task ->
// release) and polls the task to a terminal state. It is the key real-hardware
// validation for the firmware protocol variance.
func runFirmwareCheck(h *harness, o options) {
	enabled, reason := o.gate(o.firmwareFile != "")
	if !enabled {
		h.record("firmware.install", statusSkip, reason)
		return
	}

	f, err := os.Open(o.firmwareFile)
	if err != nil {
		h.record("firmware.install", statusFail, "open image: "+err.Error())
		return
	}
	defer f.Close()

	taskID, err := h.conn.FirmwareInstall(h.ctx, o.firmwareComponent, o.firmwareApplyTime, false, f)
	if err != nil {
		h.record("firmware.install", statusFail, "push: "+err.Error())
		return
	}
	h.record("firmware.install", statusPass, "pushed; taskID="+taskID)

	// Poll the Task resource to a terminal state.
	for {
		select {
		case <-h.ctx.Done():
			h.record("firmware.task", statusFail, "timed out polling task "+taskID+": "+h.ctx.Err().Error())
			return
		default:
		}

		state, status, terr := h.conn.FirmwareTaskStatus(h.ctx, constants.FirmwareInstallStepInstallStatus, o.firmwareComponent, taskID, "")
		if terr != nil {
			h.record("firmware.task", statusFail, "poll: "+terr.Error())
			return
		}

		fmt.Printf("       firmware task %s: state=%s %s\n", taskID, state, status)

		switch state {
		case constants.Complete:
			h.record("firmware.task", statusPass, "task completed: "+status)
			return
		case constants.Failed:
			h.record("firmware.task", statusFail, "task failed: "+status)
			return
		}

		time.Sleep(o.firmwarePollEvery)
	}
}

// ---------------------------------------------------------------------------
// virtual-media boot flow (formerly the lenovo-vmedia-boot example)
// ---------------------------------------------------------------------------

// runVMediaBoot performs the ordered virtual-media boot flow against the XCC:
// mount the image, set it as the next boot device, then power on. The order
// matters (insert -> set next boot -> power), which is why it is a dedicated
// mode rather than relying on the fixed order of runWriteChecks. Like every
// other mutation it is gated behind -allow-writes.
func runVMediaBoot(h *harness, o options) {
	ctx := h.ctx
	c := h.conn

	if !o.allowWrites {
		h.record("vmediaboot", statusSkip, "-allow-writes not set")
		return
	}
	if o.vmediaURL == "" {
		h.record("vmediaboot", statusFail, "-vmedia-boot requires -vmedia-url")
		return
	}

	// Step 1: mount the remote image into the virtual media slot. A non-empty
	// URL inserts (ejecting anything already in the slot first); the
	// TransferProtocolType is derived from the URL scheme.
	h.action("vmediaboot.mount", true, "", func() error {
		ok, err := c.SetVirtualMedia(ctx, o.vmediaKind, o.vmediaURL)
		if err == nil && !ok {
			return fmt.Errorf("set virtual media returned ok=false")
		}
		return err
	})

	// Step 2: set the next boot device to match the mounted media kind. With
	// -bootdev-persistent it boots the media every time; otherwise it is a
	// one-time override. -bootdev-efi selects UEFI vs legacy.
	bootDevice := "cdrom"
	switch o.vmediaKind {
	case "Floppy":
		bootDevice = "floppy"
	case "USBStick":
		bootDevice = "usb"
	}
	h.action("vmediaboot.bootdev", true, "", func() error {
		ok, err := c.BootDeviceSet(ctx, bootDevice, o.bootPersistent, o.bootEFI)
		if err == nil && !ok {
			return fmt.Errorf("boot device set returned ok=false")
		}
		return err
	})

	// Step 3: power the server on (or cycle/reset). -power "" skips this.
	if o.power == "" {
		h.record("vmediaboot.power", statusSkip, "-power \"\" (pass -power on to boot now)")
		return
	}
	h.action("vmediaboot.power", true, "", func() error {
		ok, err := c.PowerSet(ctx, o.power)
		if err == nil && !ok {
			return fmt.Errorf("power %q returned ok=false", o.power)
		}
		return err
	})
}

// ---------------------------------------------------------------------------
// parse helpers
// ---------------------------------------------------------------------------

// parseKV parses "Key=Value,Key2=Value2" into a map.
func parseKV(s string) (map[string]string, error) {
	out := map[string]string{}
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("invalid key=value pair: %q", pair)
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no key=value pairs parsed from %q", s)
	}
	return out, nil
}

// parseEnable parses "enable"/"disable" (and on/off, true/false) into a bool.
func parseEnable(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "enable", "enabled", "on", "true":
		return true, nil
	case "disable", "disabled", "off", "false":
		return false, nil
	default:
		return false, fmt.Errorf("expected enable|disable, got %q", s)
	}
}

// parsePowerCap parses a watt value or "off" (nil => disable capping).
func parsePowerCap(s string) (*float64, error) {
	if strings.EqualFold(strings.TrimSpace(s), "off") {
		return nil, nil
	}
	var watts float64
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%f", &watts); err != nil {
		return nil, fmt.Errorf("invalid power cap %q: %w", s, err)
	}
	return &watts, nil
}

// validBootDevices are the bmclib boot-device names BootDeviceSet accepts.
var validBootDevices = []string{
	"bios", "cdrom", "diag", "floppy", "disk", "none",
	"pxe", "remote_drive", "sd_card", "usb", "utilities",
}

// normalizeBootDevice validates the -bootdev value against the bmclib set,
// mapping the common -vmedia-kind aliases (CD/DVD/USBStick) so users who reuse
// the virtual-media vocabulary get a working boot device instead of the opaque
// "invalid boot device" error from the wrapper.
func normalizeBootDevice(s string) (string, error) {
	d := strings.ToLower(strings.TrimSpace(s))
	switch d {
	case "cd", "dvd", "cdrom":
		return "cdrom", nil
	case "usbstick", "usb":
		return "usb", nil
	case "hdd", "harddisk":
		return "disk", nil
	}
	for _, v := range validBootDevices {
		if d == v {
			return d, nil
		}
	}
	return "", fmt.Errorf("invalid boot device %q; valid: %s", s, strings.Join(validBootDevices, " "))
}
