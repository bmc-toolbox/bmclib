# Lenovo XCC provider — per-feature hardware test plan

This document is the **per-feature test catalog** for validating the bmclib
`lenovo` provider against a real Lenovo XClarity Controller (XCC). It is the
companion to [`RUNBOOK.md`](./RUNBOOK.md): the runbook describes *how to drive
the `lenovo-smoketest` harness*; this plan defines *what to verify for every one
of the 44 advertised features*, including the mutations the harness does not yet
flag.

All unit tests run against fixtures. **Hardware is the only place the XCC
firmware/OEM variance is actually exercised** — that is what this plan is for.

---

## How to use this document

- Each feature has a test case `TC-<area>-<n>`.
- **Driver** column says how to run it:
  - `harness:-flag` → covered by `lenovo-smoketest` (see RUNBOOK §2/§3).
  - `manual:curl` / `manual:go` → not in the harness; run the raw Redfish call
    or a small Go snippet (templates in [Appendix A](#appendix-a--manual-call-templates)).
- **Independent verification** = a second, provider-independent way to confirm
  the change really took effect (XCC web UI path, or a raw Redfish GET).
- **Risk tier**:
  - 🟢 **R0 read-only** — safe on any box, including production.
  - 🟡 **R1 reversible write** — changes config; reversible; lab box only.
  - 🟠 **R2 disruptive** — reboots/resets host or BMC, or interrupts I/O.
  - 🔴 **R3 destructive / high-risk** — data loss or bricking potential
    (firmware, volume delete/init, factory reset). Dedicated lab box only,
    with a recovery plan.

A test **PASSES** when: the call returns no error, the returned/affected state
matches **Expected**, AND the **Independent verification** confirms it.

> **Instance IDs** below (`Systems/1`, `Managers/1`, `Chassis/1`,
> `Storage/RAID_Slot1`, …) are the *typical* XCC values used in the fixtures.
> On real hardware, always discover the actual IDs from the parent collection
> (`GET …/Systems`, `…/Managers`, etc.) before hand-running a `manual:curl`
> case. Pass `-v` to the harness to see the live URIs.

---

## Prerequisites

1. A lab Lenovo server with XCC reachable over HTTPS (default port 443).
2. Credentials with **Supervisor**/admin privilege (user create, BMC reset,
   firmware, factory reset all require it).
3. A second host the **XCC itself** can reach over HTTP/HTTPS/NFS for:
   - a bootable ISO (virtual media tests),
   - the SNMP/syslog trap receiver (event tests),
   - the EventService subscription destination (event subscription tests).
4. A genuine Lenovo firmware payload (`lnvgy_fw_*.uxz`) for the firmware test.
5. `curl` with `-k` (XCC ships a self-signed cert by default) for independent
   verification and `manual:curl` cases. Export once:
   ```bash
   export XCC=https://<xcc-ip>
   export CURL='curl -sk -u <user>:<pass>'      # XCC supports HTTP Basic
   ```
6. Run the **read-only sweep first** (RUNBOOK §1) and capture it with `-json`.
   Do not proceed to writes until the read sweep is `0 FAIL`.
7. To clarify exact paths/fields/AllowableValues/OEM structures before testing,
   dump every resource the provider touches (read-only):
   ```bash
   ./dump-xcc.sh <host> <user> <pass> xcc-dump
   ```
   Produces one JSON per resource + `_index.tsv` of HTTP codes (so absent
   resources like LicenseService show up as 404). With `jq` installed it also
   expands collection members and resolves the **BIOS attribute registry**
   (every valid `-set-bios` key, its type and allowable values).

---

## Coverage matrix (44 features)

| # | Feature (registry) | Test case | Driver | Risk |
|---|--------------------|-----------|--------|------|
| 1 | *(connection: Compatible/Open/Close)* | TC-FND-1 | harness (always) | 🟢 |
| 2 | *(ServiceRoot)* | TC-FND-2 | harness `serviceroot.read` | 🟢 |
| 3 | `PowerState` | TC-PWR-1 | harness `power.state` | 🟢 |
| 4 | `PowerSet` | TC-PWR-2 | harness `-power` | 🟠 |
| 5 | `BootProgress` | TC-PWR-3 | harness `boot.progress` | 🟢 |
| 6 | `BootDeviceSet` (+ override read) | TC-PWR-4 | harness `-bootdev` | 🟡 |
| 7 | `GetBiosConfiguration` | TC-BIO-1 | harness `bios.read` | 🟢 |
| 8 | `SetBiosConfiguration` | TC-BIO-2 | harness `-set-bios` | 🟡 |
| 9 | `ResetBiosConfiguration` | TC-BIO-3 | harness `-reset-bios` | 🟡 |
| 10 | `SecureBoot` (get) | TC-SEC-1 | harness `secureboot.read` | 🟢 |
| 11 | `SecureBoot` (set) | TC-SEC-2 | harness `-secureboot` | 🟡 |
| 12 | `SecureBoot` (ResetKeys) | TC-SEC-3 | manual:go | 🟡 |
| 13 | `ThermalRead` | TC-ENV-1 | harness `thermal.read` | 🟢 |
| 14 | `PowerRead` | TC-ENV-2 | harness `power.metrics` | 🟢 |
| 15 | `PowerCap` | TC-ENV-3 | harness `-powercap` | 🟡 |
| 16 | `InventoryRead` | TC-INV-1 | harness `inventory.read` | 🟢 |
| 17 | `VolumeRead` (controllers/volumes) | TC-STO-1 | harness `storage.*` | 🟢 |
| 18 | `VolumeManagement` (create) | TC-STO-2 | manual:go | 🔴 |
| 19 | `VolumeManagement` (initialize) | TC-STO-3 | manual:go | 🔴 |
| 20 | `VolumeManagement` (update) | TC-STO-4 | manual:go | 🟡 |
| 21 | `VolumeManagement` (delete) | TC-STO-5 | manual:go | 🔴 |
| 22 | `FirmwareUpload` + `FirmwareInstall*` | TC-FW-1 | harness `-firmware` | 🔴 |
| 23 | `FirmwareInstallSteps` | TC-FW-2 | harness `firmware.steps` | 🟢 |
| 24 | `FirmwareInstallStatus` / `FirmwareTaskStatus` | TC-FW-3 | harness (firmware poll) | 🟢 |
| 25 | `FirmwareInstallUploaded` / UploadInitiate | TC-FW-4 | manual:go | 🔴 |
| 26 | `SimpleUpdate` (URI install) | TC-FW-5 | manual:go | 🔴 |
| 27 | `VirtualMedia` (insert/eject) | TC-VM-1 | harness `-vmedia-url`/`-vmedia-eject` | 🟡 |
| 28 | `UnmountFloppyImage` (+ Mount=unsupported) | TC-VM-2 | manual:go | 🟡 |
| 29 | `UserRead` | TC-USR-1 | harness `users.read` | 🟢 |
| 30 | `UserCreate` | TC-USR-2 | manual:go | 🟡 |
| 31 | `UserUpdate` | TC-USR-3 | manual:go | 🟡 |
| 32 | `UserDelete` | TC-USR-4 | manual:go | 🟡 |
| 33 | *(Roles read / RoleCreate)* | TC-USR-5 | manual:go | 🟡 |
| 34 | `GetSystemEventLog` | TC-LOG-1 | harness `sel.read` | 🟢 |
| 35 | `GetSystemEventLogRaw` | TC-LOG-2 | manual:go | 🟢 |
| 36 | `ClearSystemEventLog` | TC-LOG-3 | harness `-clear-sel` | 🟡 |
| 37 | *(EventLog / audit log read)* | TC-LOG-4 | harness `log.audit` | 🟢 |
| 38 | `BmcReset` | TC-BMC-1 | harness `-bmc-reset` | 🟠 |
| 39 | *(ResetToFactoryDefaults)* | TC-BMC-2 | manual:go | 🔴 |
| 40 | *(UpdateManager)* | TC-BMC-3 | manual:go | 🟡 |
| 41 | `LicenseManagement` (read) | TC-BMC-4 | harness `license.read` | 🟢 |
| 42 | `LicenseManagement` (install/delete) | TC-BMC-5 | manual:go | 🟡 |
| 43 | `SecureKeyLifecycle` (get) | TC-BMC-6 | harness `sklm.read` | 🟢 |
| 44 | `SecureKeyLifecycle` (set servers) | TC-BMC-7 | manual:go | 🟡 |
| 45 | `NetworkInterfaceRead` (BMC/server) | TC-NET-1 | harness `network.bmc`/`network.server` | 🟢 |
| 46 | `NetworkInterfaceSet` (NIC) | TC-NET-2 | manual:go | 🟠 |
| 47 | `NetworkInterfaceSet` (host iface) | TC-NET-3 | manual:go | 🟡 |
| 48 | `NetworkProtocolRead` | TC-NET-4 | harness `network.protocols` | 🟢 |
| 49 | `NetworkProtocolSet` | TC-NET-5 | manual:go | 🟠 |
| 50 | `SerialRead` | TC-SER-1 | harness `serial.read` | 🟢 |
| 51 | `SerialSet` | TC-SER-2 | manual:go | 🟡 |
| 52 | `EventSubscription` (service/list read) | TC-EVT-1 | harness `event.*` | 🟢 |
| 53 | `EventSubscription` (create/delete) | TC-EVT-2 | manual:go | 🟡 |
| 54 | `EventSubscription` (SubmitTestEvent) | TC-EVT-3 | manual:go | 🟡 |
| 55 | `EventSubscription` (SetEventService) | TC-EVT-4 | manual:go | 🟡 |
| 56 | `Telemetry` (service/reports/defs read) | TC-TEL-1 | harness `telemetry.*` | 🟢 |
| 57 | `Telemetry` (single report/def read) | TC-TEL-2 | manual:go | 🟢 |
| 58 | `Telemetry` (SubmitTestMetricReport) | TC-TEL-3 | manual:go | 🟡 |
| 59 | `JobManagement` (service/list read) | TC-JOB-1 | harness `job.*` | 🟢 |
| 60 | `JobManagement` (single read) | TC-JOB-2 | manual:go | 🟢 |
| 61 | `JobManagement` (update schedule) | TC-JOB-3 | manual:go | 🟡 |
| 62 | `CertificateManagement` (locations/read) | TC-CRT-1 | harness `cert.*` | 🟢 |
| 63 | `CertificateManagement` (GenerateCSR) | TC-CRT-2 | manual:go | 🟢 |
| 64 | `CertificateManagement` (ReplaceCertificate) | TC-CRT-3 | manual:go | 🟠 |
| 65 | `CertificateManagement` (Rekey/Renew) | TC-CRT-4 | manual:go | 🟠 |
| 66 | `SNMP` (read) | TC-SNM-1 | harness `snmp.read` | 🟢 |
| 67 | `SNMP` (v1/v3 trap enable) | TC-SNM-2 | harness `-snmp-v1-trap`/`-snmp-v3-trap` | 🟡 |
| 68 | `SNMP` (SetSNMPAlertFilter) | TC-SNM-3 | manual:go | 🟡 |

> The 44 *registry features* map to more test cases than 44 because several
> features (firmware, volumes, certificates, events, SNMP, secure boot) bundle
> multiple distinct provider methods that each need their own hardware check.

---

## TC-FND — Foundation

### TC-FND-1 — Vendor compatibility + session lifecycle 🟢
- **Methods:** `Compatible`, `Open`, `Close`
- **Driver:** harness (runs unconditionally; the first two report lines)
- **Procedure:** `go run ./examples/lenovo-smoketest -host $H -user $U -pass $P`
- **Expected:** `connection.compatible PASS` (device identifies as Lenovo XCC) and
  `connection.open PASS` (session established). With `-basic`, repeat to confirm
  HTTP Basic auth also works.
- **Independent verification:** `$CURL $XCC/redfish/v1/ | jq '.Vendor,.Product'`
  → contains `Lenovo`. Check the XCC active-session count does **not** climb
  after repeated runs (session is released on Close — XCC caps at 16).
- **Pass criteria:** both PASS; non-Lenovo box correctly reported NOT compatible.

### TC-FND-2 — Service root 🟢
- **Method:** `ServiceRoot`
- **Driver:** harness `serviceroot.read`
- **Expected:** `vendor="Lenovo" product=<model> redfish=<>=1.15.0>`.
- **Independent verification:** `$CURL $XCC/redfish/v1/ | jq '.RedfishVersion'`.

---

## TC-PWR — Power & boot

### TC-PWR-1 — Power state read 🟢
- **Method:** `PowerStateGet` — **Driver:** harness `power.state`
- **Expected:** `state=on` / `off` matching the box.
- **Verify:** XCC UI → *Power* tile; or `$CURL $XCC/redfish/v1/Systems/1 | jq .PowerState`.

### TC-PWR-2 — Power set 🟠
- **Method:** `PowerSet(state)` — **Driver:** `harness -allow-writes -power <on|off|soft|cycle|reset>`
- **Procedure:** test each transition on a lab box; allow the host to settle
  between states. Note: `on` is a no-op (returns ok) when already on.
- **Expected:** `power.set PASS`; host changes state.
- **Verify:** observe POST to `…/Systems/1/Actions/ComputerSystem.Reset` (`-v`);
  confirm physical/UI power state after each call. Map: on→`On`, off→`ForceOff`,
  soft→`GracefulShutdown`, cycle→`ForceRestart`/`PowerCycle`, reset→`ForceRestart`.
- **Risk:** abrupt off/cycle/reset interrupt the OS.

### TC-PWR-3 — Boot progress 🟢
- **Method:** `GetBootProgress` — **Driver:** harness `boot.progress`
- **Note:** requires RedfishVersion ≥ 1.13.0 (XCC reports 1.15.0). WARN if the
  model omits `BootProgress`.
- **Expected:** `lastState=<SystemHardwareInitializationComplete|OSRunning|…>`.

### TC-PWR-4 — Boot device override 🟡
- **Methods:** `BootDeviceOverrideGet`, `BootDeviceSet`
- **Driver:** read = harness `boot.override`; write = `harness -allow-writes -bootdev <pxe|disk|cdrom|bios> [-bootdev-efi]`
- **Procedure:** set one-time legacy PXE, re-read; set one-time UEFI cdrom, re-read.
- **⚠️ XCC allows ONE-TIME only:** `BootSourceOverrideEnabled@Redfish.AllowableValues`
  is `{Once, Disabled}` — `Continuous` is rejected (PropertyValueNotInList).
  The provider rejects `-bootdev-persistent` up front with a clear error;
  persistent boot is managed via the `BootOrder`, not the override.
  (`-bootdev` also accepts CD/DVD/USBStick aliases → cdrom/usb.)
- **Expected:** `boot.set PASS`; subsequent `boot.override` reflects device and
  `efi`, with `persistent=false`.
- **Verify:** XCC UI → *Server Configuration → Boot Options*; or
  `$CURL $XCC/redfish/v1/Systems/1 | jq .Boot` →
  `BootSourceOverrideTarget/Enabled/Mode`. Reboot once to confirm a one-time
  override clears itself.

---

## TC-BIO — BIOS

### TC-BIO-1 — BIOS read 🟢
- **Method:** `GetBiosConfiguration` — **Driver:** harness `bios.read`
- **Expected:** `N attributes` (N > 0).
- **Verify:** `$CURL $XCC/redfish/v1/Systems/1/Bios | jq '.Attributes|length'`.

### TC-BIO-2 — BIOS set 🟡
- **Method:** `SetBiosConfiguration(map)` — **Driver:** `harness -allow-writes -set-bios "Key=Value,..."`
- **⚠️ Use a real attribute name from THIS box's registry.** Read TC-BIO-1 first
  and copy exact attribute keys — names are model/firmware-specific. On a
  ThinkSystem SR630 V2 the AMI-style `QuietBoot` does **not** exist and XCC
  returns `Base.1.11.PropertyUnknown`; `BootModes_SystemBootMode=UEFIMode` is
  valid there. Don't copy the example key blindly.
- **Apply-time note (provider fix):** XCC's BIOS settings resource **rejects**
  the `@Redfish.SettingsApplyTime`/`ApplyTime` annotation that the shared
  redfishwrapper sends. The lenovo provider overrides `SetBiosConfiguration` to
  PATCH `Attributes` only (no apply-time); XCC stages the change for the next
  host reset implicitly. (Hardware-discovered 2026-06-09 on SR630 V2.)
- **Expected:** `bios.set PASS` (provider GETs the settings target for the ETag,
  then PATCHes Attributes only).
- **Verify:** `$CURL $XCC/redfish/v1/Systems/1/Bios | jq '."@Redfish.Settings".SettingsObject."@odata.id"'`
  to find the settings target, GET it for pending `Attributes`; after a host
  reboot confirm the active value changed.

### TC-BIO-3 — BIOS reset to defaults 🟡
- **Method:** `ResetBiosConfiguration` — **Driver:** `harness -allow-writes -reset-bios`
- **Expected:** `bios.reset PASS` (POST `…/Bios/Actions/Bios.ResetBios`).
- **Verify:** UI shows a pending "restore defaults" on next boot; reboot and spot
  check a previously-changed attribute returned to default.

---

## TC-SEC — Secure boot

### TC-SEC-1 — Secure boot read 🟢
- **Method:** `GetSecureBoot` — **Driver:** harness `secureboot.read`
- **Expected:** `enabled=<bool> mode=<> current=<>`.
- **Verify:** `$CURL $XCC/redfish/v1/Systems/1/SecureBoot | jq`.

### TC-SEC-2 — Secure boot enable/disable 🟡
- **Method:** `SetSecureBoot(bool)` — **Driver:** `harness -allow-writes -secureboot <enable|disable>`
- **Expected:** `secureboot.set PASS`; re-read shows new `enabled`.
- **Verify:** UI → *Security → Secure Boot*; effective after host reboot.

### TC-SEC-3 — Reset secure-boot keys 🟡 *(not in harness)*
- **Method:** `ResetSecureBootKeys(resetType)` — **Driver:** `manual:go`
- **Procedure:** call with `resetType="ResetAllKeysToDefault"` (also valid:
  `DeleteAllKeys`, `DeletePK`). See [Appendix A](#appendix-a--manual-call-templates).
- **Expected:** no error; POST to `…/SecureBoot/Actions/SecureBoot.ResetKeys`.
- **Verify:** re-read secure boot; UI key store reflects the reset.

---

## TC-ENV — Thermal & power metrics

### TC-ENV-1 — Thermal read 🟢
- **Method:** `Thermal` — **Driver:** harness `thermal.read`
- **Expected:** `N temps, M fans` (both > 0 on a powered box).
- **Verify:** `$CURL $XCC/redfish/v1/Chassis/1/Thermal | jq '.Temperatures|length,.Fans|length'`.

### TC-ENV-2 — Power metrics 🟢
- **Method:** `ReadPower` — **Driver:** harness `power.metrics`
- **Expected:** `consumed=<W> capacity=<W> supplies=<N>`.
- **Verify:** `$CURL $XCC/redfish/v1/Chassis/1/Power | jq '.PowerControl,.PowerSupplies|length'`.

### TC-ENV-3 — Power cap set/clear 🟡
- **Method:** `SetPowerCap(*float64)` — **Driver:** `harness -allow-writes -powercap <watts|off>`
- **Procedure:** set a cap above current draw (e.g. `-powercap 1200`), re-read
  `power.metrics`, then `-powercap off` to disable.
- **Expected:** `power.cap PASS`; cap reflected then cleared.
- **Verify:** `$CURL $XCC/redfish/v1/Chassis/1/Power | jq '.PowerControl[0].PowerLimit'`;
  UI → *Power → Power Capping*.

---

## TC-INV — Inventory

### TC-INV-1 — Hardware inventory 🟢
- **Method:** `Inventory` (delegates to redfishwrapper) — **Driver:** harness `inventory.read`
- **Expected:** `vendor="Lenovo" model=<> cpus>0 dimms>0 drives>=0 nics>0`.
- **Note:** record the XCC **firmware version + model** from this line for the
  release-gate evidence.
- **Verify:** spot-check counts against UI → *Inventory*. If `WithFailInventoryOnError`
  is set, a single component error fails the whole read — test both modes.

---

## TC-STO — Storage & volumes

### TC-STO-1 — Storage controllers + volumes read 🟢
- **Methods:** `StorageControllers`, `Volumes(storageID)` — **Driver:** harness `storage.controllers` / `storage.volumes`
- **Expected:** `N controllers`; for the first controller, `M volumes`.
- **Verify:** `$CURL $XCC/redfish/v1/Systems/1/Storage | jq '.Members'` then a
  controller's `/Volumes`.

> ⚠️ **TC-STO-2..5 are 🔴 destructive RAID operations.** Run only on a lab box
> with **no data you care about**. Always capture controller + volume state
> before and after, and have the RAID layout documented for restore.

### TC-STO-2 — Volume create 🔴 *(not in harness)*
- **Method:** `VolumeCreate(storageID, VolumeCreateRequest)` — **Driver:** `manual:go`
- **Procedure:** build a `bmc.VolumeCreateRequest` (RAID type, drive IDs, capacity)
  from drives discovered in TC-STO-1; call `VolumeCreate`.
- **Expected:** returns a non-empty volume ID (from `Location` header or body `Id`).
- **Verify:** `GET …/Storage/<ctrl>/Volumes` lists the new volume; UI → *Storage*.

### TC-STO-3 — Volume initialize 🔴 *(not in harness)*
- **Method:** `VolumeInitialize(storageID, volumeID, initType)` — `initType` ∈ `Fast|Slow`
- **Expected:** no error; POST `…/Volumes/<id>/Actions/Volume.Initialize` with only `InitializeType`.
- **Verify:** volume `Status`/operation shows initializing → OK in UI.

### TC-STO-4 — Volume update 🟡 *(not in harness)*
- **Method:** `VolumeUpdate(storageID, volumeID, settings map)`
- **Procedure:** PATCH a benign property (e.g. volume name / write-cache policy).
- **Verify:** re-read the volume; property changed.

### TC-STO-5 — Volume delete 🔴 *(not in harness)*
- **Method:** `VolumeDelete(storageID, volumeID)`
- **Expected:** no error; volume gone from the collection.
- **Verify:** `GET …/Volumes` no longer lists it; UI confirms.

---

## TC-FW — Firmware (🔴 dedicated lab box only)

### TC-FW-1 — Push-protocol firmware install 🔴
- **Methods:** `FirmwareInstall` (claim → push `/mfwupdate` or `/fwupdate` → poll Task → release)
- **Driver:** `harness -allow-writes -firmware <file> -firmware-component <id> -firmware-applytime <Immediate|OnReset|OnStartUpdateRequest> -timeout 30m`
- **Procedure:** see RUNBOOK §3. Use a real `lnvgy_fw_*.uxz`; prefer a backup
  target (`-firmware-component BMC-Backup`) for the first run.
- **Expected:** `firmware.install PASS` (pushed; taskID=…) then `firmware.task PASS`
  (task completed). The harness **polls the Task resource, never the
  TaskMonitor URI** (a GET on TaskMonitor deletes a finished task on XCC).
- **Verify:** UI → *Firmware Update* shows the new version after apply;
  `HttpPushUriTargetsBusy` returns to `false` (claim released). Confirm a second
  install can start (busy flag not stuck).
- **Pass criteria:** task reaches `Complete`; version updated; busy released even
  on a failed/aborted push.

### TC-FW-2 — Install step plan (read) 🟢
- **Method:** `FirmwareInstallSteps(component)` — **Driver:** harness `firmware.steps`
- **Expected:** `N install steps` (the ordered plan: upload → install status …).

### TC-FW-3 — Task status polling 🟢
- **Methods:** `FirmwareTaskStatus`, `FirmwareInstallStatus`
- **Driver:** exercised inside TC-FW-1's poll loop; can also be called standalone
  against a known task ID.
- **Expected:** maps XCC task state → `initializing|running|complete|failed`.
- **Verify:** cross-check against `$CURL $XCC/redfish/v1/TaskService/Tasks/<id>`.

### TC-FW-4 — Upload-then-install (two-phase) 🔴 *(not in harness)*
- **Methods:** `FirmwareUpload` / `FirmwareInstallUploadAndInitiate` → `FirmwareInstallUploaded(component, uploadTaskID)`
- **Driver:** `manual:go` — upload returns a task id; pass it to the uploaded-install call.
- **Use when:** validating `OnStartUpdateRequest` apply-time / staged installs.

### TC-FW-5 — SimpleUpdate (install from URI) 🔴 *(not in harness)*
- **Method:** `SimpleUpdate(imageURI, transferProtocol)` — **Driver:** `manual:go`
- **Procedure:** host the image on a reachable HTTP/TFTP server; call with the
  URI and protocol (`HTTP`/`HTTPS`/`TFTP`).
- **Expected:** POST `…/UpdateService/Actions/UpdateService.SimpleUpdate`; task id returned.
- **Verify:** poll the task (TC-FW-3) to completion; version updated.

---

## TC-VM — Virtual media

### TC-VM-1 — Insert / eject virtual media 🟡
- **Method:** `SetVirtualMedia(kind, url)` (insert) / `SetVirtualMedia(kind, "")` (eject)
- **Driver:** `harness -allow-writes -vmedia-url <http://host/boot.iso> -vmedia-kind CD`
  then `harness -allow-writes -vmedia-eject -vmedia-kind CD`
- **Critical hardware note:** real XCC slots (EXT1-4/Remote1-4/RDOC1-2) **do NOT
  expose the Redfish `InsertMedia`/`EjectMedia` actions** — the provider inserts
  via **PATCH** on the VirtualMedia resource (`Image/Inserted/WriteProtected/
  TransferProtocolType`). This is the single most important XCC-variance test.
- **Procedure:**
  1. Insert ISO → confirm exactly **one** slot occupied.
  2. Insert again (same kind) → must **replace**, not create a second mount.
  3. Eject → slot cleared.
- **Expected:** each call PASS; re-mount does not accumulate mounts.
- **Verify:** `$CURL $XCC/redfish/v1/Managers/1/VirtualMedia/<slot> | jq '.Image,.Inserted'`;
  UI → *Remote Console → Mount Media*. Optionally boot the host from the mounted ISO.

### TC-VM-2 — Floppy mount/unmount 🟡 *(mount = unsupported by design)*
- **Methods:** `MountFloppyImage` (returns "unsupported"), `UnmountFloppyImage`
- **Driver:** `manual:go`
- **Expected:** `MountFloppyImage` returns the explicit *unsupported* error (XCC
  has no raw floppy upload in its Redfish API — by design, not a bug);
  `UnmountFloppyImage` ejects the Floppy slot (idempotent if none inserted).
- **Verify:** Floppy slot `Inserted=false` after unmount.

---

## TC-USR — Users, accounts, roles

### TC-USR-1 — User read 🟢
- **Method:** `UserRead` — **Driver:** harness `users.read`
- **Expected:** `N accounts` (includes the built-in USERID/Administrator).
- **Verify:** `$CURL $XCC/redfish/v1/AccountService/Accounts | jq '.Members|length'`.

### TC-USR-2 — User create 🟡 *(not in harness)*
- **Method:** `UserCreate(user, pass, role)` — **Driver:** `manual:go`
- **Procedure:** create `smoketest` with role `Administrator` (or `Operator`).
  Provider tries Redfish `CreateAccount` first, then falls back to PATCH of an
  empty slot (Intel-Purley style) — both paths should be exercised if possible.
- **Expected:** no error; new account appears.
- **Verify:** `GET …/Accounts`; **log in as the new user** to confirm the
  password/role actually work (don't trust the POST alone).

### TC-USR-3 — User update 🟡 *(not in harness)*
- **Method:** `UserUpdate(user, pass, role)`
- **Procedure:** change the `smoketest` user's password and role.
- **Verify:** log in with the new password; confirm new privilege.

### TC-USR-4 — User delete 🟡 *(not in harness)*
- **Method:** `UserDelete(user)`
- **Expected:** account removed; login as that user now fails.
- **Verify:** `GET …/Accounts` no longer lists it. **Clean up the test user here.**

### TC-USR-5 — Roles read + custom role create 🟡 *(not in harness)*
- **Methods:** `Roles`, `RoleCreate(roleID, privileges)`
- **Expected:** `Roles` lists predefined roles; `RoleCreate` POSTs a custom role
  to `…/AccountService/Roles`.
- **Verify:** `GET …/AccountService/Roles` lists the new role with the privileges.
- **Note (DEFERRED):** AccountService lockout/LDAP read+PATCH is **not
  implemented** (OpenSpec task 2.3, no spec scenario) — out of scope for this gate.

---

## TC-LOG — Logs & SEL

### TC-LOG-1 — SEL read (parsed) 🟢
- **Method:** `GetSystemEventLog` — **Driver:** harness `sel.read`
- **Note:** aggregates **all** manager log services (Sel + AuditLog) → may return
  more than just SEL rows.
- **Verify:** `$CURL $XCC/redfish/v1/Managers/1/LogServices/Sel/Entries | jq '.Members|length'`.

### TC-LOG-2 — SEL read (raw) 🟢 *(not in harness)*
- **Method:** `GetSystemEventLogRaw` — **Driver:** `manual:go`
- **Expected:** raw entries payload returned.

### TC-LOG-3 — SEL clear 🟡
- **Method:** `ClearSystemEventLog` — **Driver:** `harness -allow-writes -clear-sel`
- **Caveat:** SEL get reads via *Managers* path; clear acts via *Chassis*
  `…/LogServices/Sel/Actions/LogService.ClearLog` (wrapper design).
- **Expected:** `sel.clear PASS`.
- **Verify:** re-read SEL → entry count drops (UI → *Event Log*).

### TC-LOG-4 — Audit / OEM log services read 🟢
- **Method:** `EventLog(ctx, logServiceID)` — **Driver:** harness `log.audit`
  (`lenovo.LogServiceAudit`)
- **Other IDs:** `LogServiceActive`, `LogServicePlatform`, `LogServiceMaintenance`,
  `LogServiceServiceAdvisor`, `LogServiceDiagnostic` — run `manual:go` for each
  the model exposes.

---

## TC-BMC — BMC management

### TC-BMC-1 — BMC reset 🟠
- **Method:** `BmcReset(resetType)` — **Driver:** `harness -allow-writes -bmc-reset <GracefulRestart|ForceRestart>`
- **Expected:** `bmc.reset PASS`; **the session drops** as the BMC restarts.
- **Verify:** XCC pings back after ~1–3 min; web UI reachable again. Host power
  state is unaffected.

### TC-BMC-2 — Reset to factory defaults 🔴 *(not in harness)*
- **Method:** `ResetToFactoryDefaults(resetType)` — `resetType` default `ResetAll`
- **Driver:** `manual:go` — **last test on the box**; wipes BMC config (network,
  users, certs). Have console/physical access to re-IP afterward.
- **Verify:** BMC returns to defaults (DHCP/default creds per the chosen scope).

### TC-BMC-3 — Update manager config 🟡 *(not in harness)*
- **Method:** `UpdateManager(properties map)`
- **Procedure:** PATCH a benign OEM/timezone property (e.g. `DateTimeLocalOffset`).
- **Verify:** `GET …/Managers/1` shows the new value; UI → *BMC Configuration*.

### TC-BMC-4 — License read 🟢
- **Method:** `Licenses` — **Driver:** harness `license.read`
- **Verify:** `$CURL $XCC/redfish/v1/LicenseService/Licenses | jq '.Members'`.
- **Hardware finding (2026-06-09, SR630 V2):** this firmware returned **404
  RequestUriNotFound** for `/redfish/v1/LicenseService/Licenses`. The provider
  now treats an absent LicenseService (404) as **"no licenses"** — `Licenses`
  returns an empty slice with no error, so `license.read` reports **PASS
  (0 licenses)** instead of WARN. Other errors still propagate. FOLLOW-UP
  (deferred): confirm whether such models expose licenses under a different path
  (e.g. `Managers/<id>/Oem/Lenovo`) and read from there instead of stubbing.

### TC-BMC-5 — License install / delete 🟡 *(not in harness)*
- **Methods:** `LicenseInstall(string)`, `LicenseDelete(id)` — **Driver:** `manual:go`
- **Procedure:** install a valid XCC license string (e.g. XCC Advanced trial),
  confirm, then delete it.
- **Expected:** install adds a license member; delete removes it.
- **Verify:** UI → *BMC Configuration → License* before/after.

### TC-BMC-6 — SKLM read 🟢
- **Method:** `GetSecureKeyLifecycle` — **Driver:** harness `sklm.read`
- **Expected:** `deviceGroup=<> keyServers=<N>` (WARN if SKLM not licensed/present).

### TC-BMC-7 — SKLM set key-repo servers 🟡 *(not in harness)*
- **Method:** `SetSecureKeyRepoServers([]SecureKeyRepoServer)` — **Driver:** `manual:go`
- **Procedure:** PATCH a test `HostName`/`Port` list to
  `…/Managers/1/Oem/Lenovo/SecureKeyLifecycleService`.
- **Verify:** re-read SKLM config shows the servers.

---

## TC-NET — Network & TC-SER — Serial

### TC-NET-1 — NIC read (BMC + server) 🟢
- **Methods:** `BMCNetworkInterfaces`, `ServerNetworkInterfaces` — **Driver:** harness `network.bmc`/`network.server`
- **Verify:** `$CURL $XCC/redfish/v1/Managers/1/EthernetInterfaces` and
  `…/Systems/1/EthernetInterfaces`.

### TC-NET-2 — BMC NIC set 🟠 *(not in harness)*
- **Method:** `SetBMCNetworkInterface(id, attrs)` — **Driver:** `manual:go`
- **⚠️ Self-lockout risk:** changing the BMC IP/VLAN can cut your connection.
  Test a non-management property first, or have console access.
- **Verify:** re-read the interface; UI → *Network*.

### TC-NET-3 — Host interface enable 🟡 *(not in harness)*
- **Method:** `SetHostInterfaceEnabled(id, enabled)`
- **Verify:** `GET …/Managers/1/HostInterfaces/<id>` → `InterfaceEnabled`.

### TC-NET-4 — Network protocols read 🟢
- **Method:** `NetworkProtocols` — **Driver:** harness `network.protocols`
- **Expected:** services list (HTTP/HTTPS/SSH/IPMI/SNMP/NTP/KVMIP/VirtualMedia/SSDP/Telnet) with port + enabled.
- **Verify:** `$CURL $XCC/redfish/v1/Managers/1/NetworkProtocol | jq`.

### TC-NET-5 — Network protocols set 🟠 *(not in harness)*
- **Method:** `SetNetworkProtocols(attrs)` — **Driver:** `manual:go`
- **⚠️** Disabling HTTPS/SSH can lock you out. Toggle a low-risk one (e.g. SSDP)
  and revert.
- **Verify:** re-read protocols; the toggled service changed.

### TC-SER-1 — Serial read 🟢
- **Method:** `SerialInterfaces` — **Driver:** harness `serial.read`
- **Verify:** `$CURL $XCC/redfish/v1/Managers/1/SerialInterfaces | jq '.Members'`.

### TC-SER-2 — Serial set 🟡 *(not in harness)*
- **Method:** `SetSerialInterface(id, attrs)` — **Driver:** `manual:go`
- **Procedure:** PATCH `BitRate`/`Parity` on a serial interface and revert.
- **Verify:** re-read shows the new value.

---

## TC-EVT — Events & TC-TEL — Telemetry

### TC-EVT-1 — Event service + subscriptions read 🟢
- **Methods:** `EventService`, `EventSubscriptions` — **Driver:** harness `event.service`/`event.subscriptions`
- **Verify:** `$CURL $XCC/redfish/v1/EventService` and `…/EventService/Subscriptions`.

### TC-EVT-2 — Subscription create / delete 🟡 *(not in harness)*
- **Methods:** `EventSubscriptionCreate(req)`, `EventSubscriptionDelete(id)` — **Driver:** `manual:go`
- **Procedure:** create a subscription with `Destination` = a reachable HTTP
  listener you control; capture the returned id; delete it.
- **Expected:** create returns a non-empty id (from `Location`); delete removes it.
- **Verify:** `GET …/EventService/Subscriptions` before/after; if you also run
  TC-EVT-3, confirm your listener receives the test event.

### TC-EVT-3 — Submit test event 🟡 *(not in harness)*
- **Method:** `SubmitTestEvent(messageID)` — **Driver:** `manual:go`
- **Verify:** the subscribed destination receives the event payload.

### TC-EVT-4 — Set event service 🟡 *(not in harness)*
- **Method:** `SetEventService(attrs)` — PATCH e.g. `ServiceEnabled` / delivery-retry.
- **Verify:** re-read EventService.

### TC-TEL-1 — Telemetry service + reports + definitions read 🟢
- **Methods:** `TelemetryService`, `MetricReports`, `MetricReportDefinitions`, `MetricDefinitions`
- **Driver:** harness `telemetry.service`/`telemetry.reports`/`telemetry.definitions`
- **Expected:** WARN acceptable if the model has no TelemetryService.

### TC-TEL-2 — Single metric report / definition read 🟢 *(not in harness)*
- **Methods:** `MetricReport(id)`, `MetricReportDefinitions` (then `MetricReport` per id)
- **Verify:** `$CURL $XCC/redfish/v1/TelemetryService/MetricReports/<id>`.

### TC-TEL-3 — Submit test metric report 🟡 *(not in harness)*
- **Method:** `SubmitTestMetricReport(reportName)` — **Driver:** `manual:go`
- **Verify:** POST `…/TelemetryService.SubmitTestMetricReport` succeeds; report appears.

---

## TC-JOB — Jobs

### TC-JOB-1 — Job service + list read 🟢
- **Methods:** `JobService`, `Jobs` — **Driver:** harness `job.service`/`job.list`

### TC-JOB-2 — Single job read 🟢 *(not in harness)*
- **Method:** `Job(id)` — read a real job id from TC-JOB-1.

### TC-JOB-3 — Update job schedule 🟡 *(not in harness)*
- **Method:** `JobUpdateSchedule(id, schedule)` — **Driver:** `manual:go`
- **Verify:** re-read the job; schedule changed.

---

## TC-CRT — Certificates

### TC-CRT-1 — Certificate locations + read 🟢
- **Methods:** `CertificateLocations`, `Certificate(location)` — **Driver:** harness `cert.locations`/`cert.read`
- **Verify:** `$CURL $XCC/redfish/v1/CertificateService/CertificateLocations`.

### TC-CRT-2 — Generate CSR 🟢 *(not in harness)*
- **Method:** `GenerateCSR(CSRRequest)` — **Driver:** `manual:go`
- **Expected:** returns a PEM CSR string.
- **Verify:** `openssl req -in csr.pem -noout -text` parses the subject you sent.

### TC-CRT-3 — Replace certificate 🟠 *(not in harness)*
- **Method:** `ReplaceCertificate(certificatePEM, targetURI)`
- **⚠️** Replacing the HTTPS cert resets TLS — the browser/clients will see the
  new cert; an invalid PEM can break HTTPS access. Test with a valid signed cert.
- **Verify:** browser shows the new cert; `openssl s_client -connect <xcc>:443`.

### TC-CRT-4 — Rekey / renew certificate 🟠 *(not in harness)*
- **Methods:** `RekeyCertificate(certURI)`, `RenewCertificate(certURI)`
- **Verify:** new cert serial/validity on the HTTPS endpoint.

---

## TC-SNM — SNMP

### TC-SNM-1 — SNMP read 🟢
- **Method:** `SNMP` — **Driver:** harness `snmp.read`
- **Expected:** `v1Trap=<> v3Trap=<> port=<>`.

### TC-SNM-2 — Enable/disable v1 / v3 trap 🟡
- **Methods:** `EnableSNMPv1Trap(bool)`, `EnableSNMPv3Trap(bool)`
- **Driver:** `harness -allow-writes -snmp-v1-trap <enable|disable>` / `-snmp-v3-trap …`
- **Verify:** re-read SNMP; UI → *BMC Configuration → SNMP*; a configured trap
  receiver gets a test trap if you also fire one.

### TC-SNM-3 — Set SNMP alert filter 🟡 *(not in harness)*
- **Method:** `SetSNMPAlertFilter(attrs)` — **Driver:** `manual:go`
- **Verify:** re-read; PATCH applied to `…/NetworkProtocol/Oem/Lenovo/SNMP`.

---

## Appendix A — manual-call templates

For `manual:go` cases, drop a throwaway `main()` next to the harness (or extend
it with a new flag). Skeleton — connect like the harness, then call the method:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/bmc-toolbox/bmclib/v2/bmc"
    "github.com/bmc-toolbox/bmclib/v2/providers/lenovo"
    logrusr "github.com/bombsimon/logrusr/v2"
    "github.com/sirupsen/logrus"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    logger := logrusr.New(logrus.New())
    c := lenovo.New("<host>", "<user>", "<pass>", logger, lenovo.WithPort("443"))
    if !c.Compatible(ctx) { panic("not a Lenovo XCC") }
    if err := c.Open(ctx); err != nil { panic(err) }
    defer c.Close(ctx)

    // --- pick the call under test, e.g.: ---

    // TC-SEC-3
    // err := c.ResetSecureBootKeys(ctx, "ResetAllKeysToDefault")

    // TC-USR-2 / 3 / 4
    // err := c.UserCreate(ctx, "smoketest", "S3cret!Passw0rd", "Administrator")
    // err := c.UserUpdate(ctx, "smoketest", "New!Passw0rd", "Operator")
    // err := c.UserDelete(ctx, "smoketest")

    // TC-USR-5
    // err := c.RoleCreate(ctx, "custom", []string{"Login", "ConfigureManager"})
    // roles, err := c.Roles(ctx)

    // TC-STO-2..5
    // id, err := c.VolumeCreate(ctx, "<storageID>", bmc.VolumeCreateRequest{ /* RAIDType, Drives, CapacityBytes */ })
    // err := c.VolumeInitialize(ctx, "<storageID>", id, "Fast")
    // err := c.VolumeUpdate(ctx, "<storageID>", id, map[string]any{"Name": "newname"})
    // err := c.VolumeDelete(ctx, "<storageID>", id)

    // TC-FW-5
    // taskID, err := c.SimpleUpdate(ctx, "http://10.0.0.2/lnvgy_fw.uxz", "HTTP")

    // TC-BMC-2 / 3 / 5 / 7
    // err := c.ResetToFactoryDefaults(ctx, "ResetAll")   // LAST — wipes config
    // err := c.UpdateManager(ctx, map[string]any{"DateTimeLocalOffset": "+00:00"})
    // err := c.LicenseInstall(ctx, "<base64-license>"); err = c.LicenseDelete(ctx, "<id>")
    // err := c.SetSecureKeyRepoServers(ctx, []bmc.SecureKeyRepoServer{{HostName: "kmip.lab", Port: 5696}})

    // TC-NET-2 / 3 / 5, TC-SER-2
    // err := c.SetBMCNetworkInterface(ctx, "eth0", map[string]any{ /* ... */ })
    // err := c.SetHostInterfaceEnabled(ctx, "1", true)
    // err := c.SetNetworkProtocols(ctx, map[string]any{"SSDP": map[string]any{"ProtocolEnabled": false}})
    // err := c.SetSerialInterface(ctx, "1", map[string]any{"BitRate": "115200"})

    // TC-EVT-2 / 3 / 4
    // id, err := c.EventSubscriptionCreate(ctx, bmc.EventSubscriptionRequest{ /* Destination, Protocol, RegistryPrefixes */ })
    // err := c.SubmitTestEvent(ctx, "<MessageId>")
    // err := c.EventSubscriptionDelete(ctx, id)
    // err := c.SetEventService(ctx, map[string]any{"ServiceEnabled": true})

    // TC-TEL-3, TC-SNM-3, TC-CRT-2..4, TC-LOG-2
    // err := c.SubmitTestMetricReport(ctx, "<reportName>")
    // err := c.SetSNMPAlertFilter(ctx, map[string]any{ /* ... */ })
    // csr, err := c.GenerateCSR(ctx, bmc.CSRRequest{ /* CommonName, Org, ... */ })
    // err := c.ReplaceCertificate(ctx, "<pem>", "/redfish/v1/Managers/1/NetworkProtocol/HTTPS/Certificates/1")
    // err := c.RekeyCertificate(ctx, "<certURI>"); err = c.RenewCertificate(ctx, "<certURI>")
    // entries, err := c.GetSystemEventLogRaw(ctx)

    fmt.Println("done")
}
```

Raw Redfish independent verification (no provider): every read above has a
`$CURL $XCC/redfish/v1/...` form shown inline. XCC accepts HTTP Basic on all
resources except it requires auth beyond `/redfish/v1/`.

---

## Release-gate sign-off

Record per the RUNBOOK "What to record" list, plus this table:

| TC | Run? | Result (PASS/WARN/FAIL/NA) | Evidence (log / UI screenshot / curl) | Notes |
|----|------|----------------------------|---------------------------------------|-------|
| TC-FND-1 … TC-SNM-3 | | | | |

**Gate to ship:** every 🟢 read = PASS or justified WARN; every 🟡/🟠 mutation
exercised on the lab box = PASS with independent verification; TC-FW-1 (firmware
push) + TC-VM-1 (vmedia PATCH) = PASS (these are the two highest-variance XCC
behaviours the fixtures cannot prove). 🔴 destructive cases (TC-STO-2/3/5,
TC-BMC-2) run last, on a throwaway box, with a documented recovery path.
