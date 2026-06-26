# Lenovo XCC provider — real-hardware smoke test runbook

`lenovo-smoketest` validates the bmclib `lenovo` provider against a real Lenovo
XClarity Controller (XCC). The unit tests run entirely against fixtures; this
harness is the bridge to actual hardware.

## Safety model

- **Read-only by default.** A plain invocation performs only GET/read calls and
  never changes the BMC or host.
- **Mutations are double-gated.** A mutating operation runs only when BOTH the
  master `-allow-writes` flag AND that operation's own flag are given. Otherwise
  it is reported as `SKIP`.
- **Exit code.** `0` when there are no `FAIL` results, `1` otherwise. `WARN`
  (a capability the model does not expose) does not fail the run.

## Result statuses

| Status | Meaning |
|--------|---------|
| `PASS` | the check succeeded |
| `WARN` | a non-core read failed — the XCC model may not expose it (telemetry, jobs, SKLM, audit log, …) |
| `FAIL` | a core capability failed (compatibility, service root, power state, inventory) or a requested mutation failed |
| `SKIP` | a mutation whose flag (or `-allow-writes`) was not provided |

## 1. Read-only sweep (run this first)

```bash
go run ./examples/lenovo-smoketest -host <xcc-ip> -user <user> -pass <pass>
# add -v for HTTP debug logging, -json to capture the report for CI
```

Expected: a list of `[PASS]`/`[WARN]`/`[SKIP]` lines and a summary with
`0 FAIL`. Capture the output. Investigate every `FAIL`; review `WARN`s to
confirm they correspond to genuinely-absent capabilities on the model under
test.

Reads exercised: serviceroot, power state, boot progress + override, BIOS,
secure boot, thermal, power metrics, inventory, storage controllers + volumes,
users, SEL, audit log, licenses, SKLM, BMC/server NICs, network protocols,
serial, event service + subscriptions, telemetry service + reports + definitions,
job service + list, certificate locations + read, SNMP, firmware install steps.

## 2. Targeted mutation checks (opt-in, one at a time)

Run against a lab/non-production server. Each needs `-allow-writes` plus its flag.

```bash
# power (use a state safe for the box under test)
go run ./examples/lenovo-smoketest -host <ip> -user <u> -pass <p> -allow-writes -power on

# next-boot device (one-time override). NOTE: XCC allows only a one-time (Once)
# override — it rejects -bootdev-persistent (Continuous); manage persistent boot
# via the BootOrder instead.
... -allow-writes -bootdev pxe                 # one-time, legacy
... -allow-writes -bootdev cdrom -bootdev-efi  # one-time, UEFI (boot a mounted ISO)

# BIOS set / reset
# BIOS attribute NAMES are model/firmware-specific. Discover the exact keys on
# THIS box first (one unknown key fails the whole PATCH with PropertyUnknown):
#     go run ./examples/lenovo-smoketest -host <ip> -user <u> -pass <p> -v 2>/dev/null | grep "bios attr"
# then set a key copied verbatim from that list, e.g. on a ThinkSystem SR630 V2:
... -allow-writes -set-bios "BootModes_SystemBootMode=UEFIMode"
... -allow-writes -reset-bios

# secure boot / power cap
... -allow-writes -secureboot enable
... -allow-writes -powercap 1200       # or: -powercap off

# virtual media (host the ISO somewhere reachable by the XCC)
... -allow-writes -vmedia-url http://10.0.0.2/boot.iso -vmedia-kind CD
... -allow-writes -vmedia-eject -vmedia-kind CD

# logs / SNMP / BMC reset / NMI
... -allow-writes -clear-sel
... -allow-writes -snmp-v1-trap enable
... -allow-writes -bmc-reset GracefulRestart    # disconnects the session
... -allow-writes -nmi
```

## 2b. Virtual-media boot flow (ordered: mount → set boot → power)

A focused mode (folded in from the former `lenovo-vmedia-boot` example) that
mounts an image, points the next boot at it, and powers the host on — in that
order. It replaces the read/write sweep and is gated by `-allow-writes`.

```bash
go run ./examples/lenovo-smoketest -host <ip> -user <u> -pass <p> \
    -allow-writes -vmedia-boot \
    -vmedia-url http://10.0.0.2/installer.iso -vmedia-kind CD \
    -bootdev-efi -power on
```

Notes: the boot device is derived from `-vmedia-kind` (CD/DVD→`cdrom`,
Floppy→`floppy`, USBStick→`usb`); the override is one-time (XCC rejects a
persistent/`Continuous` override — `-bootdev-persistent` will error); pass
`-power on` explicitly (empty `-power` mounts + sets boot but skips powering on).

## 3. Firmware install (highest-risk — dedicated lab box only)

The firmware path validates the XCC push protocol (claim `HttpPushUriTargetsBusy`
→ multipart `/mfwupdate` or raw `/fwupdate` → poll the Task → release). This is
the protocol variance the unit tests can only fixture-test.

```bash
go run ./examples/lenovo-smoketest -host <ip> -user <u> -pass <p> \
    -allow-writes \
    -firmware /path/to/lnvgy_fw_*.uxz \
    -firmware-component BMC-Backup \
    -firmware-applytime OnReset \
    -timeout 30m -firmware-poll 15s
```

The harness pushes the image, prints the task id, then polls the Task resource
(never the TaskMonitor URI) until `complete`/`failed`. Use a long `-timeout`.

## What to record for the release gate

1. The full read-only sweep output (`-json` recommended) with `0 FAIL`.
2. XCC firmware version / model (from the `inventory.read` line).
3. For each mutation exercised: the `PASS` line and independent confirmation on
   the BMC UI/CLI that the change took effect.
4. The firmware task transcript (`pushed; taskID=…` → `task completed`).
5. Any `WARN`/`FAIL` with notes on whether it is expected for the model.
