#!/usr/bin/env bash
#
# dump-xcc.sh — READ-ONLY Redfish dumper for the Lenovo XCC bmclib provider.
#
# GETs every Redfish resource the `lenovo` provider touches and saves the raw
# JSON under an output directory, so you can verify against real hardware the
# paths, field names, OEM structures, action targets and *@Redfish.AllowableValues
# the provider assumes. It performs ONLY GETs — it never mutates the BMC.
#
# Usage:
#   ./dump-xcc.sh <host> <user> <pass> [outdir]
#   XCC_INSECURE=0 ./dump-xcc.sh ...     # verify TLS (default: -k, insecure)
#   TOP=5 ./dump-xcc.sh ...              # cap log-entry pages at $top=5 (default 5)
#
# Requires: curl. Optional: jq (enables collection-member expansion + pretty
# printing + the BIOS attribute registry resolve; without it you get top-level
# resources only). Output: one .json per resource + an _index.tsv of HTTP codes.
#
set -u

HOST="${1:-}"; USERNAME="${2:-}"; PASSWORD="${3:-}"; OUTDIR="${4:-xcc-dump}"
if [[ -z "$HOST" || -z "$USERNAME" || -z "$PASSWORD" ]]; then
  echo "usage: $0 <host> <user> <pass> [outdir]" >&2
  exit 2
fi

XCC="https://${HOST}"
TOP="${TOP:-5}"
CURL_INSECURE="-k"; [[ "${XCC_INSECURE:-1}" == "0" ]] && CURL_INSECURE=""
AUTH=(-u "${USERNAME}:${PASSWORD}")
HAVE_JQ=0; command -v jq >/dev/null 2>&1 && HAVE_JQ=1

mkdir -p "$OUTDIR"
INDEX="$OUTDIR/_index.tsv"
: > "$INDEX"
[[ $HAVE_JQ -eq 0 ]] && echo "WARN: jq not found — collections will not be expanded and output is raw" >&2

# get <path> — GET $XCC<path>, save pretty JSON to a file named after the path,
# and record "HTTP_CODE\tpath" in the index. Safe to call on absent resources.
get() {
  local path="$1"
  local name="${path#/redfish/v1/}"; name="${name//\//__}"; name="${name%__}"
  [[ -z "$name" ]] && name="serviceroot"
  local out="$OUTDIR/${name}.json"
  local code
  code=$(curl -s $CURL_INSECURE "${AUTH[@]}" -H 'Accept: application/json' \
               --connect-timeout 10 --max-time 60 \
               -o "${out}.raw" -w '%{http_code}' "${XCC}${path}")
  printf '%s\t%s\n' "$code" "$path" | tee -a "$INDEX"
  if [[ ! -s "${out}.raw" ]]; then
    # No body (connection failure, or an empty response): leave a placeholder.
    rm -f "${out}.raw"; : > "$out"; return 0
  fi
  if [[ $HAVE_JQ -eq 1 ]] && jq . "${out}.raw" >"$out" 2>/dev/null; then
    rm -f "${out}.raw"
  else
    mv -f "${out}.raw" "$out"
  fi
}

# members <path> — print each Members[].@odata.id of a collection (needs jq).
members() {
  [[ $HAVE_JQ -eq 0 ]] && return 0
  local out="$OUTDIR/$(echo "${1#/redfish/v1/}" | sed 's#/#__#g; s#__$##').json"
  [[ -f "$out" ]] || return 0
  jq -r '.Members[]?."@odata.id" // empty' "$out" 2>/dev/null
}

# get_members <collection-path> — dump the collection, then every member.
get_members() { get "$1"; local m; for m in $(members "$1"); do get "$m"; done; }

section() { echo; echo "### $1"; }

# --------------------------------------------------------------------------
section "Service root, registries, metadata"
get "/redfish/v1/"
get_members "/redfish/v1/Registries"
# Resolve and dump the BIOS attribute registry (all valid attribute names,
# types and AllowableValues — the source of truth for -set-bios keys).
if [[ $HAVE_JQ -eq 1 && -f "$OUTDIR/Registries.json" ]]; then
  for r in $(members "/redfish/v1/Registries"); do
    rid=$(basename "$r")
    case "$rid" in
      *BiosAttributeRegistry*|*LenovoBiosAttributeRegistry*)
        get "$r"
        loc=$(jq -r '.Location[]?.Uri // empty' "$OUTDIR/$(echo "${r#/redfish/v1/}" | sed 's#/#__#g').json" 2>/dev/null | head -1)
        [[ -n "$loc" ]] && get "$loc"
        ;;
    esac
  done
fi

section "Systems: power, boot, BIOS, secure boot, NICs"
get_members "/redfish/v1/Systems"
sys=$(members "/redfish/v1/Systems" | head -1); sys="${sys:-/redfish/v1/Systems/1}"
get "${sys}/Bios"
get "${sys}/Bios/Pending"                       # @Redfish.Settings target (set-bios)
get "${sys}/Bios/SD"                             # alt settings target on some FW
get "${sys}/Bios/ChangePasswordActionInfo"       # ChangePassword allowable values
get "${sys}/SecureBoot"
get "${sys}/ResetActionInfo"                      # ComputerSystem.Reset allowable values
get_members "${sys}/EthernetInterfaces"

section "Chassis: thermal, power metrics, power cap"
get_members "/redfish/v1/Chassis"
ch=$(members "/redfish/v1/Chassis" | head -1); ch="${ch:-/redfish/v1/Chassis/1}"
get "${ch}/Thermal"
get "${ch}/Power"

section "Storage and volumes"
get_members "${sys}/Storage"
for st in $(members "${sys}/Storage"); do
  get_members "${st}/Volumes"
  get_members "${st}/Drives"
done

section "Managers: BMC, network, serial, vmedia, SKLM, SNMP, logs"
get_members "/redfish/v1/Managers"
mgr=$(members "/redfish/v1/Managers" | head -1); mgr="${mgr:-/redfish/v1/Managers/1}"
get "${mgr}/NetworkProtocol"
get "${mgr}/NetworkProtocol/Oem/Lenovo/SNMP"
get "${mgr}/Oem/Lenovo/SecureKeyLifecycleService"
get_members "${mgr}/EthernetInterfaces"
get_members "${mgr}/HostInterfaces"
get_members "${mgr}/SerialInterfaces"
get_members "${mgr}/VirtualMedia"
# HTTPS certificate collection (cert replace/rekey/renew targets live here)
get_members "${mgr}/NetworkProtocol/HTTPS/Certificates"

section "Log services (SEL, audit, ...) — first \$top=$TOP entries each"
get_members "${mgr}/LogServices"
for ls in $(members "${mgr}/LogServices"); do
  get "${ls}"
  get "${ls}/Entries?\$top=${TOP}"
done
get_members "${ch}/LogServices"
for ls in $(members "${ch}/LogServices"); do get "${ls}"; done

section "Firmware: update service, inventory, tasks"
get "/redfish/v1/UpdateService"
get_members "/redfish/v1/UpdateService/FirmwareInventory"
get_members "/redfish/v1/TaskService/Tasks"
get "/redfish/v1/TaskService"

section "Accounts and roles"
get "/redfish/v1/AccountService"
get_members "/redfish/v1/AccountService/Accounts"
get_members "/redfish/v1/AccountService/Roles"

section "Events and telemetry"
get "/redfish/v1/EventService"
get_members "/redfish/v1/EventService/Subscriptions"
get "/redfish/v1/TelemetryService"
get_members "/redfish/v1/TelemetryService/MetricReports"
get_members "/redfish/v1/TelemetryService/MetricReportDefinitions"
get_members "/redfish/v1/TelemetryService/MetricDefinitions"

section "Jobs"
get "/redfish/v1/JobService"
get_members "/redfish/v1/JobService/Jobs"

section "Certificates"
get "/redfish/v1/CertificateService"
get "/redfish/v1/CertificateService/CertificateLocations"

section "License (404 on firmware levels without LicenseService)"
get "/redfish/v1/LicenseService"
get_members "/redfish/v1/LicenseService/Licenses"

echo
echo "done. dumps in: $OUTDIR/   index: $INDEX"
echo "non-200 responses (absent/locked resources):"
awk -F'\t' '$1!=200 {printf "  %s  %s\n", $1, $2}' "$INDEX" || true
