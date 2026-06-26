package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// This file provides the FromInterfaces dispatchers for the extended
// vendor-capability interfaces (secure boot, thermal, power capping, storage
// volumes, licensing, secure key lifecycle, networking, serial, events,
// telemetry, jobs, certificates and SNMP).
//
// To avoid repeating the provider-iteration/timeout/metadata boilerplate that
// the older hand-written *FromInterfaces functions carry, the dispatch is
// centralised in two generics: runProviderRead (operations returning a value)
// and runProviderAction (operations returning only an error). Each public
// *FromInterfaces function below is a thin, type-safe binding over one of them.

// runProviderRead invokes op on the first registered provider that implements
// interface T, returning its result. It records provider attempts, the
// successful provider and per-provider failures in the returned Metadata, and
// applies the per-provider timeout — matching the behaviour of the original
// hand-written *FromInterfaces helpers.
func runProviderRead[T any, R any](
	ctx context.Context,
	timeout time.Duration,
	providers []interface{},
	ifaceName string,
	op func(ctx context.Context, impl T) (R, error),
) (result R, metadata Metadata, err error) {
	metadata = newMetadata()

	var matched bool
	for _, elem := range providers {
		if elem == nil {
			continue
		}

		impl, ok := elem.(T)
		if !ok {
			continue
		}
		matched = true

		name := getProviderName(elem)

		select {
		case <-ctx.Done():
			return result, metadata, multierror.Append(err, ctx.Err())
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, name)

			callCtx, cancel := context.WithTimeout(ctx, timeout)
			res, opErr := op(callCtx, impl)
			cancel()

			if opErr != nil {
				err = multierror.Append(err, errors.WithMessagef(opErr, "provider: %v", name))
				metadata.FailedProviderDetail[name] = opErr.Error()
				continue
			}

			metadata.SuccessfulProvider = name

			return res, metadata, nil
		}
	}

	if !matched {
		return result, metadata, multierror.Append(err, fmt.Errorf("no %s implementations found", ifaceName))
	}

	return result, metadata, multierror.Append(err, fmt.Errorf("failed to invoke %s", ifaceName))
}

// runProviderAction is runProviderRead for operations that return only an error.
func runProviderAction[T any](
	ctx context.Context,
	timeout time.Duration,
	providers []interface{},
	ifaceName string,
	op func(ctx context.Context, impl T) error,
) (Metadata, error) {
	_, metadata, err := runProviderRead(ctx, timeout, providers, ifaceName,
		func(ctx context.Context, impl T) (struct{}, error) {
			return struct{}{}, op(ctx, impl)
		})

	return metadata, err
}

// --- SecureBootManager ---

func GetSecureBootFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) (SecureBootState, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "SecureBootManager",
		func(ctx context.Context, p SecureBootManager) (SecureBootState, error) { return p.GetSecureBoot(ctx) })
}

func SetSecureBootFromInterfaces(ctx context.Context, timeout time.Duration, enabled bool, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "SecureBootManager",
		func(ctx context.Context, p SecureBootManager) error { return p.SetSecureBoot(ctx, enabled) })
}

func ResetSecureBootKeysFromInterfaces(ctx context.Context, timeout time.Duration, resetType string, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "SecureBootManager",
		func(ctx context.Context, p SecureBootManager) error { return p.ResetSecureBootKeys(ctx, resetType) })
}

// --- ThermalReader ---

func ThermalFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) (ThermalReading, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "ThermalReader",
		func(ctx context.Context, p ThermalReader) (ThermalReading, error) { return p.Thermal(ctx) })
}

// --- PowerReader / PowerCapSetter ---

func ReadPowerFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) (PowerInfo, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "PowerReader",
		func(ctx context.Context, p PowerReader) (PowerInfo, error) { return p.ReadPower(ctx) })
}

func SetPowerCapFromInterfaces(ctx context.Context, timeout time.Duration, limitWatts *float64, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "PowerCapSetter",
		func(ctx context.Context, p PowerCapSetter) error { return p.SetPowerCap(ctx, limitWatts) })
}

// --- VolumeManager ---

func StorageControllersFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]StorageControllerInfo, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "VolumeManager",
		func(ctx context.Context, p VolumeManager) ([]StorageControllerInfo, error) {
			return p.StorageControllers(ctx)
		})
}

func VolumesFromInterfaces(ctx context.Context, timeout time.Duration, storageID string, providers []interface{}) ([]VolumeInfo, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "VolumeManager",
		func(ctx context.Context, p VolumeManager) ([]VolumeInfo, error) { return p.Volumes(ctx, storageID) })
}

func VolumeCreateFromInterfaces(ctx context.Context, timeout time.Duration, storageID string, req VolumeCreateRequest, providers []interface{}) (string, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "VolumeManager",
		func(ctx context.Context, p VolumeManager) (string, error) { return p.VolumeCreate(ctx, storageID, req) })
}

func VolumeInitializeFromInterfaces(ctx context.Context, timeout time.Duration, storageID, volumeID, initType string, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "VolumeManager",
		func(ctx context.Context, p VolumeManager) error {
			return p.VolumeInitialize(ctx, storageID, volumeID, initType)
		})
}

func VolumeUpdateFromInterfaces(ctx context.Context, timeout time.Duration, storageID, volumeID string, settings map[string]any, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "VolumeManager",
		func(ctx context.Context, p VolumeManager) error {
			return p.VolumeUpdate(ctx, storageID, volumeID, settings)
		})
}

func VolumeDeleteFromInterfaces(ctx context.Context, timeout time.Duration, storageID, volumeID string, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "VolumeManager",
		func(ctx context.Context, p VolumeManager) error { return p.VolumeDelete(ctx, storageID, volumeID) })
}

// --- LicenseManager ---

func LicensesFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]LicenseInfo, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "LicenseManager",
		func(ctx context.Context, p LicenseManager) ([]LicenseInfo, error) { return p.Licenses(ctx) })
}

func LicenseInstallFromInterfaces(ctx context.Context, timeout time.Duration, license string, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "LicenseManager",
		func(ctx context.Context, p LicenseManager) error { return p.LicenseInstall(ctx, license) })
}

func LicenseDeleteFromInterfaces(ctx context.Context, timeout time.Duration, licenseID string, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "LicenseManager",
		func(ctx context.Context, p LicenseManager) error { return p.LicenseDelete(ctx, licenseID) })
}

// --- SecureKeyLifecycle ---

func GetSecureKeyLifecycleFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) (SecureKeyLifecycleConfig, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "SecureKeyLifecycle",
		func(ctx context.Context, p SecureKeyLifecycle) (SecureKeyLifecycleConfig, error) {
			return p.GetSecureKeyLifecycle(ctx)
		})
}

func SetSecureKeyRepoServersFromInterfaces(ctx context.Context, timeout time.Duration, servers []SecureKeyRepoServer, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "SecureKeyLifecycle",
		func(ctx context.Context, p SecureKeyLifecycle) error { return p.SetSecureKeyRepoServers(ctx, servers) })
}

// --- NetworkInterfaceGetter / Setter ---

func BMCNetworkInterfacesFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]NetworkInterface, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "NetworkInterfaceGetter",
		func(ctx context.Context, p NetworkInterfaceGetter) ([]NetworkInterface, error) {
			return p.BMCNetworkInterfaces(ctx)
		})
}

func ServerNetworkInterfacesFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]NetworkInterface, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "NetworkInterfaceGetter",
		func(ctx context.Context, p NetworkInterfaceGetter) ([]NetworkInterface, error) {
			return p.ServerNetworkInterfaces(ctx)
		})
}

func SetBMCNetworkInterfaceFromInterfaces(ctx context.Context, timeout time.Duration, id string, attrs map[string]any, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "NetworkInterfaceSetter",
		func(ctx context.Context, p NetworkInterfaceSetter) error {
			return p.SetBMCNetworkInterface(ctx, id, attrs)
		})
}

func SetHostInterfaceEnabledFromInterfaces(ctx context.Context, timeout time.Duration, id string, enabled bool, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "NetworkInterfaceSetter",
		func(ctx context.Context, p NetworkInterfaceSetter) error {
			return p.SetHostInterfaceEnabled(ctx, id, enabled)
		})
}

// --- NetworkProtocolGetter / Setter ---

func NetworkProtocolsFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]NetworkProtocol, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "NetworkProtocolGetter",
		func(ctx context.Context, p NetworkProtocolGetter) ([]NetworkProtocol, error) {
			return p.NetworkProtocols(ctx)
		})
}

func SetNetworkProtocolsFromInterfaces(ctx context.Context, timeout time.Duration, attrs map[string]any, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "NetworkProtocolSetter",
		func(ctx context.Context, p NetworkProtocolSetter) error { return p.SetNetworkProtocols(ctx, attrs) })
}

// --- SerialInterfaceGetter / Setter ---

func SerialInterfacesFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]SerialInterface, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "SerialInterfaceGetter",
		func(ctx context.Context, p SerialInterfaceGetter) ([]SerialInterface, error) {
			return p.SerialInterfaces(ctx)
		})
}

func SetSerialInterfaceFromInterfaces(ctx context.Context, timeout time.Duration, id string, attrs map[string]any, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "SerialInterfaceSetter",
		func(ctx context.Context, p SerialInterfaceSetter) error { return p.SetSerialInterface(ctx, id, attrs) })
}

// --- EventSubscriber ---

func EventServiceFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) (EventServiceInfo, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "EventSubscriber",
		func(ctx context.Context, p EventSubscriber) (EventServiceInfo, error) { return p.EventService(ctx) })
}

func EventSubscriptionsFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]EventSubscription, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "EventSubscriber",
		func(ctx context.Context, p EventSubscriber) ([]EventSubscription, error) {
			return p.EventSubscriptions(ctx)
		})
}

func EventSubscriptionCreateFromInterfaces(ctx context.Context, timeout time.Duration, req EventSubscriptionRequest, providers []interface{}) (string, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "EventSubscriber",
		func(ctx context.Context, p EventSubscriber) (string, error) {
			return p.EventSubscriptionCreate(ctx, req)
		})
}

func EventSubscriptionDeleteFromInterfaces(ctx context.Context, timeout time.Duration, id string, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "EventSubscriber",
		func(ctx context.Context, p EventSubscriber) error { return p.EventSubscriptionDelete(ctx, id) })
}

func SubmitTestEventFromInterfaces(ctx context.Context, timeout time.Duration, messageID string, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "EventSubscriber",
		func(ctx context.Context, p EventSubscriber) error { return p.SubmitTestEvent(ctx, messageID) })
}

func SetEventServiceFromInterfaces(ctx context.Context, timeout time.Duration, attrs map[string]any, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "EventSubscriber",
		func(ctx context.Context, p EventSubscriber) error { return p.SetEventService(ctx, attrs) })
}

// --- TelemetryReader ---

func TelemetryServiceFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) (TelemetryServiceInfo, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "TelemetryReader",
		func(ctx context.Context, p TelemetryReader) (TelemetryServiceInfo, error) {
			return p.TelemetryService(ctx)
		})
}

func MetricReportsFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]string, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "TelemetryReader",
		func(ctx context.Context, p TelemetryReader) ([]string, error) { return p.MetricReports(ctx) })
}

func MetricReportFromInterfaces(ctx context.Context, timeout time.Duration, id string, providers []interface{}) (MetricReportInfo, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "TelemetryReader",
		func(ctx context.Context, p TelemetryReader) (MetricReportInfo, error) { return p.MetricReport(ctx, id) })
}

func MetricReportDefinitionsFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]string, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "TelemetryReader",
		func(ctx context.Context, p TelemetryReader) ([]string, error) { return p.MetricReportDefinitions(ctx) })
}

func MetricDefinitionsFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]string, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "TelemetryReader",
		func(ctx context.Context, p TelemetryReader) ([]string, error) { return p.MetricDefinitions(ctx) })
}

func SubmitTestMetricReportFromInterfaces(ctx context.Context, timeout time.Duration, reportName string, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "TelemetryReader",
		func(ctx context.Context, p TelemetryReader) error { return p.SubmitTestMetricReport(ctx, reportName) })
}

// --- JobManager ---

func JobServiceFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) (JobServiceInfo, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "JobManager",
		func(ctx context.Context, p JobManager) (JobServiceInfo, error) { return p.JobService(ctx) })
}

func JobsFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]JobInfo, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "JobManager",
		func(ctx context.Context, p JobManager) ([]JobInfo, error) { return p.Jobs(ctx) })
}

func JobFromInterfaces(ctx context.Context, timeout time.Duration, id string, providers []interface{}) (JobInfo, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "JobManager",
		func(ctx context.Context, p JobManager) (JobInfo, error) { return p.Job(ctx, id) })
}

func JobUpdateScheduleFromInterfaces(ctx context.Context, timeout time.Duration, id string, schedule map[string]any, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "JobManager",
		func(ctx context.Context, p JobManager) error { return p.JobUpdateSchedule(ctx, id, schedule) })
}

// --- CertificateManager ---

func CertificateLocationsFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) ([]string, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "CertificateManager",
		func(ctx context.Context, p CertificateManager) ([]string, error) { return p.CertificateLocations(ctx) })
}

func CertificateFromInterfaces(ctx context.Context, timeout time.Duration, location string, providers []interface{}) (CertificateInfo, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "CertificateManager",
		func(ctx context.Context, p CertificateManager) (CertificateInfo, error) {
			return p.Certificate(ctx, location)
		})
}

func GenerateCSRFromInterfaces(ctx context.Context, timeout time.Duration, req CSRRequest, providers []interface{}) (string, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "CertificateManager",
		func(ctx context.Context, p CertificateManager) (string, error) { return p.GenerateCSR(ctx, req) })
}

func ReplaceCertificateFromInterfaces(ctx context.Context, timeout time.Duration, certificatePEM, targetURI string, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "CertificateManager",
		func(ctx context.Context, p CertificateManager) error {
			return p.ReplaceCertificate(ctx, certificatePEM, targetURI)
		})
}

func RekeyCertificateFromInterfaces(ctx context.Context, timeout time.Duration, certURI string, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "CertificateManager",
		func(ctx context.Context, p CertificateManager) error { return p.RekeyCertificate(ctx, certURI) })
}

func RenewCertificateFromInterfaces(ctx context.Context, timeout time.Duration, certURI string, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "CertificateManager",
		func(ctx context.Context, p CertificateManager) error { return p.RenewCertificate(ctx, certURI) })
}

// --- SNMPConfigurer ---

func SNMPFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) (SNMPConfig, Metadata, error) {
	return runProviderRead(ctx, timeout, providers, "SNMPConfigurer",
		func(ctx context.Context, p SNMPConfigurer) (SNMPConfig, error) { return p.SNMP(ctx) })
}

func SetSNMPAlertFilterFromInterfaces(ctx context.Context, timeout time.Duration, attrs map[string]any, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "SNMPConfigurer",
		func(ctx context.Context, p SNMPConfigurer) error { return p.SetSNMPAlertFilter(ctx, attrs) })
}

func EnableSNMPv1TrapFromInterfaces(ctx context.Context, timeout time.Duration, enabled bool, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "SNMPConfigurer",
		func(ctx context.Context, p SNMPConfigurer) error { return p.EnableSNMPv1Trap(ctx, enabled) })
}

func EnableSNMPv3TrapFromInterfaces(ctx context.Context, timeout time.Duration, enabled bool, providers []interface{}) (Metadata, error) {
	return runProviderAction(ctx, timeout, providers, "SNMPConfigurer",
		func(ctx context.Context, p SNMPConfigurer) error { return p.EnableSNMPv3Trap(ctx, enabled) })
}
