package lenovo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/go-logr/logr"
)

const fixturesDir = "./fixtures/v1"

// testServer is an httptest-backed XCC Redfish mock. It serves recorded JSON
// fixtures and emulates Redfish session create/delete so the provider can be
// exercised entirely offline.
//
// It records whether a session was created and deleted so tests can assert the
// connection lifecycle (e.g. that Close always releases the session).
type testServer struct {
	*httptest.Server

	mu             sync.Mutex
	sessionCreated bool
	sessionDeleted bool
	// lastResetType records the ResetType posted to ComputerSystem.Reset.
	lastResetType string
	// biosReset records whether the Bios.ResetBios action was posted.
	biosReset bool
	// systemPatched records whether the ComputerSystem was PATCHed (boot set).
	systemPatched bool
	// biosPatched records whether the Bios settings target was PATCHed.
	biosPatched bool
	// biosPatchBody records the decoded body of the last Bios settings PATCH so
	// tests can assert XCC-specific payload shape (Attributes-only, no
	// @Redfish.SettingsApplyTime — which XCC rejects).
	biosPatchBody map[string]any
	// secureBootPatched records whether the SecureBoot resource was PATCHed.
	secureBootPatched bool
	// secureBootKeysReset records whether the SecureBoot.ResetKeys action ran.
	secureBootKeysReset bool
	// powerPatched records whether the Power resource was PATCHed (cap set).
	powerPatched bool
	// volumeCreated records whether a volume was POSTed to a Volumes collection.
	volumeCreated bool
	// volumeInitialized records whether the Volume.Initialize action ran.
	volumeInitialized bool
	// volumeUpdated records whether a Volume was PATCHed.
	volumeUpdated bool
	// volumeDeleted records whether a Volume was DELETEd.
	volumeDeleted bool
	// claimedBusy records whether UpdateService was PATCHed with busy=true.
	claimedBusy bool
	// releasedBusy records whether UpdateService was PATCHed with busy=false.
	releasedBusy bool
	// multipartPushed records a POST to the multipart push URI (/mfwupdate).
	multipartPushed bool
	// rawPushed records a POST to the raw push URI (/fwupdate).
	rawPushed bool
	// simpleUpdated records a POST to UpdateService.SimpleUpdate.
	simpleUpdated bool
	// startUpdated records a POST to UpdateService.StartUpdate.
	startUpdated bool
	// cdInserted records a VirtualMedia CD insert action.
	cdInserted bool
	// floppyEjected records a VirtualMedia Floppy eject action.
	floppyEjected bool
	// vmOps records the ordered VirtualMedia PATCH operations as "Slot:insert"
	// or "Slot:eject", so tests can assert eject-before-insert.
	vmOps []string
	// accountPosted records a POST to the accounts collection.
	accountPosted bool
	// accountPatched records a PATCH to an account slot.
	accountPatched bool
	// rolePosted records a POST to the roles collection.
	rolePosted bool
	// selCleared records a LogService.ClearLog action on the chassis SEL.
	selCleared bool
	// bmcReset records a Manager.Reset action.
	bmcReset bool
	// factoryReset records a Manager.ResetToDefaults action.
	factoryReset bool
	// licenseInstalled records a POST to the License collection.
	licenseInstalled bool
	// licenseDeleted records a DELETE of a License.
	licenseDeleted bool
	// sklmPatched records a PATCH of the SecureKeyLifecycleService.
	sklmPatched bool
	// bmcEthPatched records a PATCH of a BMC ethernet interface.
	bmcEthPatched bool
	// hostIfacePatched records a PATCH of a host interface.
	hostIfacePatched bool
	// netProtoPatched records a PATCH of ManagerNetworkProtocol.
	netProtoPatched bool
	// serialPatched records a PATCH of a serial interface.
	serialPatched bool
	// subscriptionCreated/Deleted record event subscription create/delete.
	subscriptionCreated bool
	subscriptionDeleted bool
	// testEventSubmitted records a SubmitTestEvent action.
	testEventSubmitted bool
	// eventServicePatched records a PATCH of the EventService.
	eventServicePatched bool
	// testMetricSubmitted records a SubmitTestMetricReport action.
	testMetricSubmitted bool
	// jobScheduleUpdated records a PATCH of a Job's Schedule.
	jobScheduleUpdated bool
	// csrGenerated/certReplaced/certRekeyed/certRenewed record certificate actions.
	csrGenerated bool
	certReplaced bool
	certRekeyed  bool
	certRenewed  bool
	// snmpPatched records a PATCH of the OEM SNMP resource.
	snmpPatched bool
}

// fixtureBytes reads a fixture file from the fixtures dir.
func fixtureBytes(t *testing.T, file string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(fixturesDir, file))
	if err != nil {
		t.Fatalf("failed to read fixture %q: %v", file, err)
	}
	return b
}

// testServerOpts configures a testServer.
type testServerOpts struct {
	// systemFixture is the fixture file served for /redfish/v1/Systems/1.
	systemFixture string
	// rejectAuth makes the session-create endpoint return HTTP 401.
	rejectAuth bool
	// updateServiceFixture is the fixture served for /redfish/v1/UpdateService.
	updateServiceFixture string
	// rejectAccountPost makes the accounts-collection POST return HTTP 405,
	// forcing the Intel-Purley slot-PATCH fallback.
	rejectAccountPost bool
	// licenseServiceNotFound drops the LicenseService routes so the mock returns
	// 404, emulating XCC firmware levels without a LicenseService collection.
	licenseServiceNotFound bool
}

// newTestServer builds and starts a TLS mock XCC server.
func newTestServer(t *testing.T, opts testServerOpts) *testServer {
	t.Helper()

	if opts.systemFixture == "" {
		opts.systemFixture = "system.lenovo.json"
	}
	if opts.updateServiceFixture == "" {
		opts.updateServiceFixture = "updateservice.json"
	}

	ts := &testServer{}

	// path -> fixture file for plain GETs.
	routes := map[string]string{
		"/redfish/v1/":                                                "serviceroot.json",
		"/redfish/v1/Systems":                                         "systems.json",
		"/redfish/v1/Systems/1":                                       opts.systemFixture,
		"/redfish/v1/Systems/1/Bios/Pending":                          "bios.pending.json",
		"/redfish/v1/Systems/1/Bios":                                  "bios.json",
		"/redfish/v1/Systems/1/SecureBoot":                            "secureboot.json",
		"/redfish/v1/Chassis":                                         "chassis.json",
		"/redfish/v1/Chassis/1":                                       "chassis.1.json",
		"/redfish/v1/Chassis/1/Power":                                 "power.json",
		"/redfish/v1/Chassis/1/Thermal":                               "thermal.json",
		"/redfish/v1/Systems/1/Storage":                               "storage.json",
		"/redfish/v1/Systems/1/Storage/RAID_Slot1":                    "storage.raid.json",
		"/redfish/v1/Systems/1/Storage/RAID_Slot1/Volumes/1":          "volume.1.json",
		"/redfish/v1/TaskService":                                     "taskservice.json",
		"/redfish/v1/TaskService/Tasks":                               "tasks.json",
		"/redfish/v1/TaskService/Tasks/1":                             "task.1.json",
		"/redfish/v1/Managers":                                        "managers.json",
		"/redfish/v1/Managers/1":                                      "manager.1.json",
		"/redfish/v1/Managers/1/VirtualMedia":                         "managers.1.virtualmedia.json",
		"/redfish/v1/AccountService":                                  "accountservice.json",
		"/redfish/v1/AccountService/Accounts/1":                       "account.1.json",
		"/redfish/v1/AccountService/Accounts/2":                       "account.2.json",
		"/redfish/v1/AccountService/Roles/Administrator":              "role.administrator.json",
		"/redfish/v1/AccountService/Roles/Operator":                   "role.operator.json",
		"/redfish/v1/Managers/1/LogServices":                          "managers.1.logservices.json",
		"/redfish/v1/Managers/1/LogServices/Sel":                      "ls.sel.json",
		"/redfish/v1/Managers/1/LogServices/AuditLog":                 "ls.audit.json",
		"/redfish/v1/Managers/1/LogServices/Sel/Entries":              "ls.sel.entries.json",
		"/redfish/v1/Managers/1/LogServices/Sel/Entries/1":            "ls.sel.entry.1.json",
		"/redfish/v1/Managers/1/LogServices/AuditLog/Entries":         "ls.audit.entries.json",
		"/redfish/v1/Managers/1/LogServices/AuditLog/Entries/1":       "ls.audit.entry.1.json",
		"/redfish/v1/Chassis/1/LogServices":                           "chassis.1.logservices.json",
		"/redfish/v1/Chassis/1/LogServices/Sel":                       "chassis.ls.sel.json",
		"/redfish/v1/LicenseService/Licenses":                         "licenses.json",
		"/redfish/v1/LicenseService/Licenses/XCC_Advanced":            "license.xcc_advanced.json",
		"/redfish/v1/Managers/1/Oem/Lenovo/SecureKeyLifecycleService": "sklm.json",
		"/redfish/v1/Managers/1/EthernetInterfaces":                   "manager.ethernetinterfaces.json",
		"/redfish/v1/Managers/1/EthernetInterfaces/eth0":              "manager.eth0.json",
		"/redfish/v1/Managers/1/HostInterfaces":                       "manager.hostinterfaces.json",
		"/redfish/v1/Managers/1/HostInterfaces/1":                     "manager.hostinterface.1.json",
		"/redfish/v1/Managers/1/SerialInterfaces":                     "manager.serialinterfaces.json",
		"/redfish/v1/Managers/1/SerialInterfaces/1":                   "manager.serial.1.json",
		"/redfish/v1/Managers/1/NetworkProtocol":                      "networkprotocol.json",
		"/redfish/v1/Systems/1/EthernetInterfaces":                    "system.ethernetinterfaces.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/NIC.1":              "system.eth.nic1.json",
		"/redfish/v1/EventService":                                    "eventservice.json",
		"/redfish/v1/EventService/Subscriptions/1":                    "subscription.1.json",
		"/redfish/v1/TelemetryService":                                "telemetryservice.json",
		"/redfish/v1/TelemetryService/MetricReports":                  "metricreports.json",
		"/redfish/v1/TelemetryService/MetricReports/PowerMetrics":     "metricreport.power.json",
		"/redfish/v1/TelemetryService/MetricReportDefinitions":        "metricreportdefinitions.json",
		"/redfish/v1/TelemetryService/MetricDefinitions":              "metricdefinitions.json",
		"/redfish/v1/JobService":                                      "jobservice.json",
		"/redfish/v1/JobService/Jobs":                                 "jobs.json",
		"/redfish/v1/JobService/Jobs/Restart":                         "job.restart.json",
		"/redfish/v1/CertificateService":                              "certificateservice.json",
		"/redfish/v1/CertificateService/CertificateLocations":         "certificatelocations.json",
		"/redfish/v1/Managers/1/NetworkProtocol/HTTPS/Certificates/1": "certificate.1.json",
		"/redfish/v1/Managers/1/NetworkProtocol/Oem/Lenovo/SNMP":      "snmp.json",
	}

	if opts.licenseServiceNotFound {
		delete(routes, "/redfish/v1/LicenseService/Licenses")
		delete(routes, "/redfish/v1/LicenseService/Licenses/XCC_Advanced")
	}

	mux := http.NewServeMux()

	// ComputerSystem.Reset action — records the requested ResetType.
	mux.HandleFunc("/redfish/v1/Systems/1/Actions/ComputerSystem.Reset", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			ResetType string `json:"ResetType"`
		}
		if body, err := io.ReadAll(r.Body); err == nil {
			_ = json.Unmarshal(body, &payload)
		}
		ts.mu.Lock()
		ts.lastResetType = payload.ResetType
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	// Bios.ResetBios action.
	mux.HandleFunc("/redfish/v1/Systems/1/Bios/Actions/Bios.ResetBios", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.biosReset = true
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	// SecureBoot.ResetKeys action.
	mux.HandleFunc("/redfish/v1/Systems/1/SecureBoot/Actions/SecureBoot.ResetKeys", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.secureBootKeysReset = true
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	// Volumes collection: GET serves the fixture, POST creates a volume and
	// returns its Location.
	mux.HandleFunc("/redfish/v1/Systems/1/Storage/RAID_Slot1/Volumes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ts.mu.Lock()
			ts.volumeCreated = true
			ts.mu.Unlock()
			w.Header().Set("Location", "/redfish/v1/Systems/1/Storage/RAID_Slot1/Volumes/2")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"@odata.id":"/redfish/v1/Systems/1/Storage/RAID_Slot1/Volumes/2","Id":"2"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixtureBytes(t, "volumes.json"))
	})

	// Volume.Initialize action.
	mux.HandleFunc("/redfish/v1/Systems/1/Storage/RAID_Slot1/Volumes/1/Actions/Volume.Initialize", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.volumeInitialized = true
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	// UpdateService: GET serves the (variant) fixture; PATCH records the
	// HttpPushUriTargetsBusy claim/release.
	mux.HandleFunc("/redfish/v1/UpdateService", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			var payload struct {
				Busy *bool `json:"HttpPushUriTargetsBusy"`
			}
			if body, err := io.ReadAll(r.Body); err == nil {
				_ = json.Unmarshal(body, &payload)
			}
			ts.mu.Lock()
			if payload.Busy != nil {
				if *payload.Busy {
					ts.claimedBusy = true
				} else {
					ts.releasedBusy = true
				}
			}
			ts.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixtureBytes(t, opts.updateServiceFixture))
	})

	// Firmware push endpoints (note: these live outside the /redfish/v1/ tree).
	mux.HandleFunc("/mfwupdate", func(w http.ResponseWriter, r *http.Request) {
		// Drain the (multipart) body before responding, else the client may see
		// a connection reset when the server closes early.
		_, _ = io.Copy(io.Discard, r.Body)
		ts.mu.Lock()
		ts.multipartPushed = true
		ts.mu.Unlock()
		w.Header().Set("Location", "/redfish/v1/TaskService/Tasks/1")
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("/fwupdate", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		ts.mu.Lock()
		ts.rawPushed = true
		ts.mu.Unlock()
		w.Header().Set("Location", "/redfish/v1/TaskService/Tasks/1")
		w.WriteHeader(http.StatusAccepted)
	})

	// UpdateService.SimpleUpdate and UpdateService.StartUpdate actions.
	mux.HandleFunc("/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.simpleUpdated = true
		ts.mu.Unlock()
		w.Header().Set("Location", "/redfish/v1/TaskService/Tasks/1")
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("/redfish/v1/UpdateService/Actions/UpdateService.StartUpdate", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.startUpdated = true
		ts.mu.Unlock()
		w.Header().Set("Location", "/redfish/v1/TaskService/Tasks/1")
		w.WriteHeader(http.StatusAccepted)
	})

	// XCC inserts/ejects virtual media via PATCH on the VirtualMedia resource
	// (not the InsertMedia/EjectMedia actions). These per-slot handlers serve the
	// slot fixture on GET and record insert/eject on PATCH (insert => Inserted
	// true with an Image; eject => Inserted false / null Image).
	vmHandler := func(slotID, fixture string) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(fixtureBytes(t, fixture))
				return
			}

			var body struct {
				Inserted *bool       `json:"Inserted"`
				Image    interface{} `json:"Image"`
			}
			if b, err := io.ReadAll(r.Body); err == nil {
				_ = json.Unmarshal(b, &body)
			}
			op := "eject"
			if body.Inserted != nil && *body.Inserted {
				op = "insert"
			}

			ts.mu.Lock()
			ts.vmOps = append(ts.vmOps, slotID+":"+op)
			if slotID == "CD" && op == "insert" {
				ts.cdInserted = true
			}
			if slotID == "Floppy" && op == "eject" {
				ts.floppyEjected = true
			}
			ts.mu.Unlock()

			w.WriteHeader(http.StatusNoContent)
		}
	}
	mux.HandleFunc("/redfish/v1/Managers/1/VirtualMedia/CD", vmHandler("CD", "vm.cd.json"))
	mux.HandleFunc("/redfish/v1/Managers/1/VirtualMedia/Floppy", vmHandler("Floppy", "vm.floppy.json"))

	// Accounts collection: GET serves the fixture (also used by gofish for an
	// etag), POST creates an account (or returns 405 to force the slot fallback).
	mux.HandleFunc("/redfish/v1/AccountService/Accounts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if opts.rejectAccountPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			ts.mu.Lock()
			ts.accountPosted = true
			ts.mu.Unlock()
			w.Header().Set("Location", "/redfish/v1/AccountService/Accounts/2")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"@odata.id":"/redfish/v1/AccountService/Accounts/2","Id":"2","UserName":"ops","RoleId":"Operator","Enabled":true}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixtureBytes(t, "accounts.json"))
	})

	// Manager.Reset and Manager.ResetToDefaults actions.
	mux.HandleFunc("/redfish/v1/Managers/1/Actions/Manager.Reset", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.bmcReset = true
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/redfish/v1/Managers/1/Actions/Manager.ResetToDefaults", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.factoryReset = true
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	// LogService.ClearLog actions (manager + chassis SEL). ClearSystemEventLog
	// clears via the chassis log services.
	mux.HandleFunc("/redfish/v1/Managers/1/LogServices/Sel/Actions/LogService.ClearLog", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/redfish/v1/Chassis/1/LogServices/Sel/Actions/LogService.ClearLog", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.selCleared = true
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	// CertificateService actions.
	mux.HandleFunc("/redfish/v1/CertificateService/Actions/CertificateService.GenerateCSR", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		ts.mu.Lock()
		ts.csrGenerated = true
		ts.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"CSRString":"-----BEGIN CERTIFICATE REQUEST-----\nMIICFDCC\n-----END CERTIFICATE REQUEST-----"}`))
	})
	mux.HandleFunc("/redfish/v1/CertificateService/Actions/CertificateService.ReplaceCertificate", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		ts.mu.Lock()
		ts.certReplaced = true
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/redfish/v1/Managers/1/NetworkProtocol/HTTPS/Certificates/1/Actions/Certificate.Rekey", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.certRekeyed = true
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/redfish/v1/Managers/1/NetworkProtocol/HTTPS/Certificates/1/Actions/Certificate.Renew", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.certRenewed = true
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	// Event subscriptions collection: GET serves the fixture, POST creates a
	// subscription and returns its Location.
	mux.HandleFunc("/redfish/v1/EventService/Subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_, _ = io.Copy(io.Discard, r.Body)
			ts.mu.Lock()
			ts.subscriptionCreated = true
			ts.mu.Unlock()
			w.Header().Set("Location", "/redfish/v1/EventService/Subscriptions/2")
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixtureBytes(t, "subscriptions.json"))
	})

	// EventService.SubmitTestEvent and TelemetryService.SubmitTestMetricReport.
	mux.HandleFunc("/redfish/v1/EventService/Actions/EventService.SubmitTestEvent", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.testEventSubmitted = true
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/redfish/v1/TelemetryService/Actions/TelemetryService.SubmitTestMetricReport", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.testMetricSubmitted = true
		ts.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	// Roles collection: GET serves the fixture, POST creates a custom role.
	mux.HandleFunc("/redfish/v1/AccountService/Roles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			ts.mu.Lock()
			ts.rolePosted = true
			ts.mu.Unlock()
			w.Header().Set("Location", "/redfish/v1/AccountService/Roles/custom")
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixtureBytes(t, "roles.json"))
	})

	// Session collection: POST creates a session, anything else is a no-op.
	mux.HandleFunc("/redfish/v1/SessionService/Sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusOK)
			return
		}

		if opts.rejectAuth {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"Base.1.8.GeneralError","message":"login failed",` +
				`"@Message.ExtendedInfo":[{"MessageId":"Base.1.8.InsufficientPrivilege",` +
				`"Message":"The credentials provided are not authorized.","Resolution":"Provide valid credentials."}]}}`))
			return
		}

		ts.mu.Lock()
		ts.sessionCreated = true
		ts.mu.Unlock()

		w.Header().Set("X-Auth-Token", "test-token")
		w.Header().Set("Location", "/redfish/v1/SessionService/Sessions/1")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"@odata.id":"/redfish/v1/SessionService/Sessions/1","Id":"1","Name":"Session"}`))
	})

	// A created session is deleted here on Close.
	mux.HandleFunc("/redfish/v1/SessionService/Sessions/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			ts.mu.Lock()
			ts.sessionDeleted = true
			ts.mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
	})

	// Catch-all for the rest of the Redfish tree.
	//
	// GETs are served from fixtures. Writes (PATCH/POST/PUT/DELETE) on a known
	// resource are accepted with 204 and recorded, so tests can assert that a
	// mutation was issued without modelling full write semantics.
	mux.HandleFunc("/redfish/v1/", func(w http.ResponseWriter, r *http.Request) {
		file, ok := routes[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodGet {
			ts.mu.Lock()
			switch r.URL.Path {
			case "/redfish/v1/Systems/1":
				ts.systemPatched = true
			case "/redfish/v1/Systems/1/Bios", "/redfish/v1/Systems/1/Bios/Pending":
				// XCC exposes a @Redfish.Settings SettingsObject at /Bios/Pending;
				// gofish PATCHes there. Record the body either way.
				ts.biosPatched = true
				if b, err := io.ReadAll(r.Body); err == nil {
					var body map[string]any
					if json.Unmarshal(b, &body) == nil {
						ts.biosPatchBody = body
					}
				}
			case "/redfish/v1/Systems/1/SecureBoot":
				ts.secureBootPatched = true
			case "/redfish/v1/Chassis/1/Power":
				ts.powerPatched = true
			case "/redfish/v1/Systems/1/Storage/RAID_Slot1/Volumes/1":
				if r.Method == http.MethodDelete {
					ts.volumeDeleted = true
				} else {
					ts.volumeUpdated = true
				}
			case "/redfish/v1/AccountService/Accounts/1", "/redfish/v1/AccountService/Accounts/2":
				ts.accountPatched = true
			case "/redfish/v1/LicenseService/Licenses":
				ts.licenseInstalled = true
			case "/redfish/v1/LicenseService/Licenses/XCC_Advanced":
				ts.licenseDeleted = true
			case "/redfish/v1/Managers/1/Oem/Lenovo/SecureKeyLifecycleService":
				ts.sklmPatched = true
			case "/redfish/v1/Managers/1/EthernetInterfaces/eth0":
				ts.bmcEthPatched = true
			case "/redfish/v1/Managers/1/HostInterfaces/1":
				ts.hostIfacePatched = true
			case "/redfish/v1/Managers/1/NetworkProtocol":
				ts.netProtoPatched = true
			case "/redfish/v1/Managers/1/SerialInterfaces/1":
				ts.serialPatched = true
			case "/redfish/v1/EventService/Subscriptions/1":
				ts.subscriptionDeleted = true
			case "/redfish/v1/EventService":
				ts.eventServicePatched = true
			case "/redfish/v1/JobService/Jobs/Restart":
				ts.jobScheduleUpdated = true
			case "/redfish/v1/Managers/1/NetworkProtocol/Oem/Lenovo/SNMP":
				ts.snmpPatched = true
			}
			ts.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
			return
		}

		body, err := os.ReadFile(filepath.Join(fixturesDir, file))
		if err != nil {
			t.Errorf("failed to read fixture %q for %s: %v", file, r.URL.Path, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})

	ts.Server = httptest.NewTLSServer(mux)

	return ts
}

// client returns a *Conn pointed at the mock server. Extra options are appended
// after the mandatory port option.
func (ts *testServer) client(t *testing.T, opts ...Option) *Conn {
	t.Helper()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("parse mock url: %v", err)
	}

	opts = append([]Option{WithPort(u.Port())}, opts...)

	return New(u.Hostname(), "user", "pass", logr.Discard(), opts...)
}

// openedClient returns a *Conn with an established session and registers Close
// + server shutdown for cleanup.
func (ts *testServer) openedClient(t *testing.T, opts ...Option) *Conn {
	t.Helper()

	c := ts.client(t, opts...)
	if err := c.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}

	t.Cleanup(func() {
		_ = c.Close(context.Background())
		ts.Close()
	})

	return c
}

func (ts *testServer) didCreateSession() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.sessionCreated
}

func (ts *testServer) didDeleteSession() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.sessionDeleted
}

func (ts *testServer) resetType() string {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.lastResetType
}

func (ts *testServer) didResetBios() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.biosReset
}

func (ts *testServer) didPatchSystem() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.systemPatched
}

func (ts *testServer) didPatchBios() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.biosPatched
}

func (ts *testServer) didPatchSecureBoot() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.secureBootPatched
}

func (ts *testServer) didResetSecureBootKeys() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.secureBootKeysReset
}

func (ts *testServer) didPatchPower() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.powerPatched
}

func (ts *testServer) didCreateVolume() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.volumeCreated
}

func (ts *testServer) didInitializeVolume() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.volumeInitialized
}

func (ts *testServer) didUpdateVolume() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.volumeUpdated
}

func (ts *testServer) didDeleteVolume() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.volumeDeleted
}

func (ts *testServer) didClaimBusy() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.claimedBusy
}

func (ts *testServer) didReleaseBusy() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.releasedBusy
}

func (ts *testServer) didMultipartPush() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.multipartPushed
}

func (ts *testServer) didRawPush() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.rawPushed
}

func (ts *testServer) didSimpleUpdate() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.simpleUpdated
}

func (ts *testServer) didStartUpdate() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.startUpdated
}

func (ts *testServer) didInsertCD() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.cdInserted
}

func (ts *testServer) didEjectFloppy() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.floppyEjected
}

func (ts *testServer) vmOperations() []string {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return append([]string(nil), ts.vmOps...)
}

func (ts *testServer) didPostAccount() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.accountPosted
}

func (ts *testServer) didPatchAccount() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.accountPatched
}

func (ts *testServer) didPostRole() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.rolePosted
}

func (ts *testServer) didClearSEL() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.selCleared
}

func (ts *testServer) didBmcReset() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.bmcReset
}

func (ts *testServer) didFactoryReset() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.factoryReset
}

func (ts *testServer) didInstallLicense() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.licenseInstalled
}

func (ts *testServer) didDeleteLicense() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.licenseDeleted
}

func (ts *testServer) didPatchSKLM() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.sklmPatched
}

func (ts *testServer) didPatchBMCEth() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.bmcEthPatched
}

func (ts *testServer) didPatchHostIface() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.hostIfacePatched
}

func (ts *testServer) didPatchNetProto() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.netProtoPatched
}

func (ts *testServer) didPatchSerial() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.serialPatched
}

func (ts *testServer) didCreateSubscription() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.subscriptionCreated
}

func (ts *testServer) didDeleteSubscription() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.subscriptionDeleted
}

func (ts *testServer) didSubmitTestEvent() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.testEventSubmitted
}

func (ts *testServer) didPatchEventService() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.eventServicePatched
}

func (ts *testServer) didSubmitTestMetric() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.testMetricSubmitted
}

func (ts *testServer) didUpdateJobSchedule() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.jobScheduleUpdated
}

func (ts *testServer) didGenerateCSR() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.csrGenerated
}

func (ts *testServer) didReplaceCertificate() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.certReplaced
}

func (ts *testServer) didRekeyCertificate() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.certRekeyed
}

func (ts *testServer) didRenewCertificate() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.certRenewed
}

func (ts *testServer) didPatchSNMP() bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.snmpPatched
}
