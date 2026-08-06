Fixtures in this directory are real Redfish responses captured from a
Supermicro `AS-1015CS-TNR-EU` (BMC firmware AMI-based, `RedfishVersion:
1.22.2`), gathered while diagnosing why `SystemBootDeviceSet(ctx, "pxe",
false, true)` never actually resulted in a UEFI PXE boot on this hardware,
despite the call returning `ok=true`.

- `serviceroot.json`, `systems.json`, `system.1.json`: real `GET` responses.
- `patch-success.json`: the real response body returned by this BMC for the
  `PATCH /redfish/v1/Systems/1` request `SystemBootDeviceSet` issues when
  called with `("pxe", false, true)` - i.e. requesting
  `BootSourceOverrideTarget: Pxe`, `BootSourceOverrideEnabled: Once`,
  `BootSourceOverrideMode: UEFI` all in one request, per
  `boot_device.go`'s `SystemBootDeviceSet`.

`system.1.json` is what a `GET` on `/redfish/v1/Systems/1` returns *after*
that PATCH: `BootSourceOverrideMode` has silently reverted to `"Legacy"`
(and `BootSourceOverrideEnabled` to `"Disabled"`) even though the request
asked for `UEFI`/`Once`. `BootSourceOverrideTarget` is the only field that
stuck as requested (`"Pxe"`).

`system.1.after-correct-set.json` is synthetic, not a real capture - a copy
of `system.1.json` with `Boot.BootSourceOverrideEnabled/Mode/Target` set to
what a well-behaved BMC would show after actually honoring the request. Used
by the happy-path test to confirm the verification added in response to this
bug doesn't produce false negatives against a BMC that works correctly.

See `boot_device_test.go`'s `TestSystemBootDeviceSet_DetectsSupermicroSilentModeDowngrade`
and `TestSystemBootDeviceSet_HappyPath`.
