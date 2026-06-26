package bmclib

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// This file exposes the high-level bmclib.Client methods for the extended
// vendor-capability interfaces. Each method resolves the registered providers,
// dispatches to the corresponding bmc.*FromInterfaces helper using the
// per-provider timeout, stores the returned metadata and records span
// attributes — the same contract as the core Client methods (GetPowerState,
// SetBootDevice, ...).

// --- Secure Boot ---

// GetSecureBoot returns the UEFI Secure Boot state.
func (c *Client) GetSecureBoot(ctx context.Context) (bmc.SecureBootState, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "GetSecureBoot")
	defer span.End()

	state, metadata, err := bmc.GetSecureBootFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return state, err
}

// SetSecureBoot enables or disables UEFI Secure Boot (applied on next boot).
func (c *Client) SetSecureBoot(ctx context.Context, enabled bool) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetSecureBoot")
	defer span.End()

	metadata, err := bmc.SetSecureBootFromInterfaces(ctx, c.perProviderTimeout(ctx), enabled, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// ResetSecureBootKeys resets the Secure Boot key databases (destructive).
func (c *Client) ResetSecureBootKeys(ctx context.Context, resetType string) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "ResetSecureBootKeys")
	defer span.End()

	metadata, err := bmc.ResetSecureBootKeysFromInterfaces(ctx, c.perProviderTimeout(ctx), resetType, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// --- Thermal / Power ---

// GetThermal returns the device's temperature and fan readings.
func (c *Client) GetThermal(ctx context.Context) (bmc.ThermalReading, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "GetThermal")
	defer span.End()

	reading, metadata, err := bmc.ThermalFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return reading, err
}

// ReadPower returns the device's power metrics and supplies.
func (c *Client) ReadPower(ctx context.Context) (bmc.PowerInfo, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "ReadPower")
	defer span.End()

	info, metadata, err := bmc.ReadPowerFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return info, err
}

// SetPowerCap sets the chassis power limit in watts; a nil limit disables capping.
func (c *Client) SetPowerCap(ctx context.Context, limitWatts *float64) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetPowerCap")
	defer span.End()

	metadata, err := bmc.SetPowerCapFromInterfaces(ctx, c.perProviderTimeout(ctx), limitWatts, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// --- Storage volumes ---

// StorageControllers lists the system's storage controllers.
func (c *Client) StorageControllers(ctx context.Context) ([]bmc.StorageControllerInfo, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "StorageControllers")
	defer span.End()

	controllers, metadata, err := bmc.StorageControllersFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return controllers, err
}

// Volumes lists the volumes managed by a storage controller.
func (c *Client) Volumes(ctx context.Context, storageID string) ([]bmc.VolumeInfo, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "Volumes")
	defer span.End()

	volumes, metadata, err := bmc.VolumesFromInterfaces(ctx, c.perProviderTimeout(ctx), storageID, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return volumes, err
}

// VolumeCreate creates a volume on a storage controller and returns its Id.
func (c *Client) VolumeCreate(ctx context.Context, storageID string, req bmc.VolumeCreateRequest) (string, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "VolumeCreate")
	defer span.End()

	id, metadata, err := bmc.VolumeCreateFromInterfaces(ctx, c.perProviderTimeout(ctx), storageID, req, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return id, err
}

// VolumeInitialize initializes a volume.
func (c *Client) VolumeInitialize(ctx context.Context, storageID, volumeID, initType string) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "VolumeInitialize")
	defer span.End()

	metadata, err := bmc.VolumeInitializeFromInterfaces(ctx, c.perProviderTimeout(ctx), storageID, volumeID, initType, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// VolumeUpdate updates volume settings.
func (c *Client) VolumeUpdate(ctx context.Context, storageID, volumeID string, settings map[string]any) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "VolumeUpdate")
	defer span.End()

	metadata, err := bmc.VolumeUpdateFromInterfaces(ctx, c.perProviderTimeout(ctx), storageID, volumeID, settings, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// VolumeDelete deletes a volume.
func (c *Client) VolumeDelete(ctx context.Context, storageID, volumeID string) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "VolumeDelete")
	defer span.End()

	metadata, err := bmc.VolumeDeleteFromInterfaces(ctx, c.perProviderTimeout(ctx), storageID, volumeID, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// --- Licenses ---

// Licenses returns the installed BMC licenses.
func (c *Client) Licenses(ctx context.Context) ([]bmc.LicenseInfo, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "Licenses")
	defer span.End()

	licenses, metadata, err := bmc.LicensesFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return licenses, err
}

// LicenseInstall installs a license from its license string.
func (c *Client) LicenseInstall(ctx context.Context, license string) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "LicenseInstall")
	defer span.End()

	metadata, err := bmc.LicenseInstallFromInterfaces(ctx, c.perProviderTimeout(ctx), license, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// LicenseDelete removes an installed license by Id.
func (c *Client) LicenseDelete(ctx context.Context, licenseID string) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "LicenseDelete")
	defer span.End()

	metadata, err := bmc.LicenseDeleteFromInterfaces(ctx, c.perProviderTimeout(ctx), licenseID, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// --- Secure Key Lifecycle ---

// GetSecureKeyLifecycle returns the Secure Key Lifecycle configuration.
func (c *Client) GetSecureKeyLifecycle(ctx context.Context) (bmc.SecureKeyLifecycleConfig, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "GetSecureKeyLifecycle")
	defer span.End()

	cfg, metadata, err := bmc.GetSecureKeyLifecycleFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return cfg, err
}

// SetSecureKeyRepoServers replaces the configured key-repository servers.
func (c *Client) SetSecureKeyRepoServers(ctx context.Context, servers []bmc.SecureKeyRepoServer) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetSecureKeyRepoServers")
	defer span.End()

	metadata, err := bmc.SetSecureKeyRepoServersFromInterfaces(ctx, c.perProviderTimeout(ctx), servers, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// --- Network interfaces & protocols ---

// BMCNetworkInterfaces returns the BMC (manager) ethernet interfaces.
func (c *Client) BMCNetworkInterfaces(ctx context.Context) ([]bmc.NetworkInterface, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "BMCNetworkInterfaces")
	defer span.End()

	ifaces, metadata, err := bmc.BMCNetworkInterfacesFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return ifaces, err
}

// ServerNetworkInterfaces returns the server (system) ethernet interfaces.
func (c *Client) ServerNetworkInterfaces(ctx context.Context) ([]bmc.NetworkInterface, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "ServerNetworkInterfaces")
	defer span.End()

	ifaces, metadata, err := bmc.ServerNetworkInterfacesFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return ifaces, err
}

// SetBMCNetworkInterface PATCHes a BMC ethernet interface.
func (c *Client) SetBMCNetworkInterface(ctx context.Context, id string, attrs map[string]any) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetBMCNetworkInterface")
	defer span.End()

	metadata, err := bmc.SetBMCNetworkInterfaceFromInterfaces(ctx, c.perProviderTimeout(ctx), id, attrs, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// SetHostInterfaceEnabled enables or disables a host interface.
func (c *Client) SetHostInterfaceEnabled(ctx context.Context, id string, enabled bool) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetHostInterfaceEnabled")
	defer span.End()

	metadata, err := bmc.SetHostInterfaceEnabledFromInterfaces(ctx, c.perProviderTimeout(ctx), id, enabled, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// NetworkProtocols returns the manager network services and their state.
func (c *Client) NetworkProtocols(ctx context.Context) ([]bmc.NetworkProtocol, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "NetworkProtocols")
	defer span.End()

	protos, metadata, err := bmc.NetworkProtocolsFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return protos, err
}

// SetNetworkProtocols PATCHes the manager network protocols.
func (c *Client) SetNetworkProtocols(ctx context.Context, attrs map[string]any) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetNetworkProtocols")
	defer span.End()

	metadata, err := bmc.SetNetworkProtocolsFromInterfaces(ctx, c.perProviderTimeout(ctx), attrs, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// --- Serial interfaces ---

// SerialInterfaces returns the BMC serial interfaces.
func (c *Client) SerialInterfaces(ctx context.Context) ([]bmc.SerialInterface, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SerialInterfaces")
	defer span.End()

	ifaces, metadata, err := bmc.SerialInterfacesFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return ifaces, err
}

// SetSerialInterface PATCHes a BMC serial interface.
func (c *Client) SetSerialInterface(ctx context.Context, id string, attrs map[string]any) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetSerialInterface")
	defer span.End()

	metadata, err := bmc.SetSerialInterfaceFromInterfaces(ctx, c.perProviderTimeout(ctx), id, attrs, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// --- Events ---

// EventService returns the event-service configuration.
func (c *Client) EventService(ctx context.Context) (bmc.EventServiceInfo, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "EventService")
	defer span.End()

	info, metadata, err := bmc.EventServiceFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return info, err
}

// EventSubscriptions lists the event subscriptions.
func (c *Client) EventSubscriptions(ctx context.Context) ([]bmc.EventSubscription, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "EventSubscriptions")
	defer span.End()

	subs, metadata, err := bmc.EventSubscriptionsFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return subs, err
}

// EventSubscriptionCreate creates an event subscription and returns its Id.
func (c *Client) EventSubscriptionCreate(ctx context.Context, req bmc.EventSubscriptionRequest) (string, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "EventSubscriptionCreate")
	defer span.End()

	id, metadata, err := bmc.EventSubscriptionCreateFromInterfaces(ctx, c.perProviderTimeout(ctx), req, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return id, err
}

// EventSubscriptionDelete deletes an event subscription by Id.
func (c *Client) EventSubscriptionDelete(ctx context.Context, id string) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "EventSubscriptionDelete")
	defer span.End()

	metadata, err := bmc.EventSubscriptionDeleteFromInterfaces(ctx, c.perProviderTimeout(ctx), id, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// SubmitTestEvent submits a test event (optionally for a specific MessageId).
func (c *Client) SubmitTestEvent(ctx context.Context, messageID string) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SubmitTestEvent")
	defer span.End()

	metadata, err := bmc.SubmitTestEventFromInterfaces(ctx, c.perProviderTimeout(ctx), messageID, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// SetEventService PATCHes event-service properties.
func (c *Client) SetEventService(ctx context.Context, attrs map[string]any) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetEventService")
	defer span.End()

	metadata, err := bmc.SetEventServiceFromInterfaces(ctx, c.perProviderTimeout(ctx), attrs, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// --- Telemetry ---

// TelemetryService returns the telemetry-service configuration.
func (c *Client) TelemetryService(ctx context.Context) (bmc.TelemetryServiceInfo, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "TelemetryService")
	defer span.End()

	info, metadata, err := bmc.TelemetryServiceFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return info, err
}

// MetricReports lists the metric report Ids.
func (c *Client) MetricReports(ctx context.Context) ([]string, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "MetricReports")
	defer span.End()

	reports, metadata, err := bmc.MetricReportsFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return reports, err
}

// MetricReport returns a metric report and its values by Id.
func (c *Client) MetricReport(ctx context.Context, id string) (bmc.MetricReportInfo, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "MetricReport")
	defer span.End()

	report, metadata, err := bmc.MetricReportFromInterfaces(ctx, c.perProviderTimeout(ctx), id, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return report, err
}

// MetricReportDefinitions lists the metric report definition Ids.
func (c *Client) MetricReportDefinitions(ctx context.Context) ([]string, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "MetricReportDefinitions")
	defer span.End()

	defs, metadata, err := bmc.MetricReportDefinitionsFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return defs, err
}

// MetricDefinitions lists the metric definition Ids.
func (c *Client) MetricDefinitions(ctx context.Context) ([]string, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "MetricDefinitions")
	defer span.End()

	defs, metadata, err := bmc.MetricDefinitionsFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return defs, err
}

// SubmitTestMetricReport submits a test metric report.
func (c *Client) SubmitTestMetricReport(ctx context.Context, reportName string) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SubmitTestMetricReport")
	defer span.End()

	metadata, err := bmc.SubmitTestMetricReportFromInterfaces(ctx, c.perProviderTimeout(ctx), reportName, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// --- Jobs ---

// JobService returns the job-service configuration.
func (c *Client) JobService(ctx context.Context) (bmc.JobServiceInfo, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "JobService")
	defer span.End()

	info, metadata, err := bmc.JobServiceFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return info, err
}

// Jobs lists the jobs.
func (c *Client) Jobs(ctx context.Context) ([]bmc.JobInfo, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "Jobs")
	defer span.End()

	jobs, metadata, err := bmc.JobsFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return jobs, err
}

// Job returns a job by Id.
func (c *Client) Job(ctx context.Context, id string) (bmc.JobInfo, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "Job")
	defer span.End()

	job, metadata, err := bmc.JobFromInterfaces(ctx, c.perProviderTimeout(ctx), id, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return job, err
}

// JobUpdateSchedule updates a job's schedule.
func (c *Client) JobUpdateSchedule(ctx context.Context, id string, schedule map[string]any) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "JobUpdateSchedule")
	defer span.End()

	metadata, err := bmc.JobUpdateScheduleFromInterfaces(ctx, c.perProviderTimeout(ctx), id, schedule, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// --- Certificates ---

// CertificateLocations returns the @odata.id locations of installed certificates.
func (c *Client) CertificateLocations(ctx context.Context) ([]string, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "CertificateLocations")
	defer span.End()

	locs, metadata, err := bmc.CertificateLocationsFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return locs, err
}

// Certificate returns a certificate's properties by its @odata.id location.
func (c *Client) Certificate(ctx context.Context, location string) (bmc.CertificateInfo, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "Certificate")
	defer span.End()

	cert, metadata, err := bmc.CertificateFromInterfaces(ctx, c.perProviderTimeout(ctx), location, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return cert, err
}

// GenerateCSR generates a certificate-signing-request and returns the PEM CSR.
func (c *Client) GenerateCSR(ctx context.Context, req bmc.CSRRequest) (string, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "GenerateCSR")
	defer span.End()

	csr, metadata, err := bmc.GenerateCSRFromInterfaces(ctx, c.perProviderTimeout(ctx), req, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return csr, err
}

// ReplaceCertificate replaces the certificate at targetURI with the given PEM.
func (c *Client) ReplaceCertificate(ctx context.Context, certificatePEM, targetURI string) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "ReplaceCertificate")
	defer span.End()

	metadata, err := bmc.ReplaceCertificateFromInterfaces(ctx, c.perProviderTimeout(ctx), certificatePEM, targetURI, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// RekeyCertificate generates a new key pair and CSR for the certificate at certURI.
func (c *Client) RekeyCertificate(ctx context.Context, certURI string) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "RekeyCertificate")
	defer span.End()

	metadata, err := bmc.RekeyCertificateFromInterfaces(ctx, c.perProviderTimeout(ctx), certURI, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// RenewCertificate generates a CSR using the existing key for the certificate at certURI.
func (c *Client) RenewCertificate(ctx context.Context, certURI string) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "RenewCertificate")
	defer span.End()

	metadata, err := bmc.RenewCertificateFromInterfaces(ctx, c.perProviderTimeout(ctx), certURI, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// --- SNMP ---

// SNMP returns the SNMP trap configuration.
func (c *Client) SNMP(ctx context.Context) (bmc.SNMPConfig, error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SNMP")
	defer span.End()

	cfg, metadata, err := bmc.SNMPFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return cfg, err
}

// SetSNMPAlertFilter PATCHes the SNMP trap (SNMPTraps) properties.
func (c *Client) SetSNMPAlertFilter(ctx context.Context, attrs map[string]any) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetSNMPAlertFilter")
	defer span.End()

	metadata, err := bmc.SetSNMPAlertFilterFromInterfaces(ctx, c.perProviderTimeout(ctx), attrs, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// EnableSNMPv1Trap enables or disables the SNMPv1 trap.
func (c *Client) EnableSNMPv1Trap(ctx context.Context, enabled bool) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "EnableSNMPv1Trap")
	defer span.End()

	metadata, err := bmc.EnableSNMPv1TrapFromInterfaces(ctx, c.perProviderTimeout(ctx), enabled, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// EnableSNMPv3Trap enables or disables the SNMPv3 trap.
func (c *Client) EnableSNMPv3Trap(ctx context.Context, enabled bool) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "EnableSNMPv3Trap")
	defer span.End()

	metadata, err := bmc.EnableSNMPv3TrapFromInterfaces(ctx, c.perProviderTimeout(ctx), enabled, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}
