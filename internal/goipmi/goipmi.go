package goipmi

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/bougou/go-ipmi"
)

// Ipmi holds the data for an ipmi connection
type Ipmi struct {
	Username    string
	Password    string
	Host        string
	Port        int
	client      *ipmi.Client
	cipherSuite int
	log         logr.Logger
}

// Option for setting optional Ipmi values
type Option func(*Ipmi)

func WithCipherSuite(cipherSuite string) Option {
	return func(i *Ipmi) {
		cipherId, err := strconv.Atoi(cipherSuite)
		if err == nil && (0 <= cipherId && cipherId <= 19) {
			i.cipherSuite = cipherId
		}
	}
}

func WithLogger(log logr.Logger) Option {
	return func(i *Ipmi) {
		i.log = log
	}
}

// New returns a new ipmi instance
func New(username, password, host string, port int, opts ...Option) (c *Ipmi, err error) {
	cl, err := ipmi.NewClient(host, port, username, password)
	if err != nil {
		return nil, err
	}
	c = &Ipmi{
		Username:    username,
		Password:    password,
		Host:        host,
		Port:        port,
		log:         logr.Discard(),
		cipherSuite: 3,
		client:      cl,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.client.WithInterface(ipmi.InterfaceLanplus)
	c.client.WithCipherSuiteID(toCipherSuiteID(c.cipherSuite))

	return c, nil
}

// parseSystemEventLogRaw parses the raw output of the system event log. Helper
// function for GetSystemEventLog to make testing the parser easier.
func parseSystemEventLog(raw string) (entries [][]string) {
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) < 6 {
			continue
		}
		if strings.TrimSpace(parts[0]) == "SEL Record ID" {
			continue
		}
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		// ID, Timestamp (date time), Description, Message (message : assertion)
		entries = append(entries, []string{parts[0], fmt.Sprintf("%s %s", parts[1], parts[2]), parts[2], fmt.Sprintf("%s : %s", parts[3], parts[4])})
	}

	return entries
}

func toCipherSuiteID(c int) ipmi.CipherSuiteID {
	if c >= 0 && c <= 19 {
		return ipmi.CipherSuiteID(c)
	}
	return ipmi.CipherSuiteID3
}

// ensureConnected ensures the IPMI client is connected
func (i *Ipmi) ensureConnected(ctx context.Context) error {
	// For go-ipmi, we need to connect before each operation
	return i.client.Connect(ctx)
}

// PowerCycle reboots the machine via bmc
func (i *Ipmi) PowerCycle(ctx context.Context) (status bool, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	_, err = i.client.ChassisControl(ctx, ipmi.ChassisControlPowerCycle)
	if err != nil {
		return false, fmt.Errorf("chassis control failed: %v", err)
	}
	return true, nil
}

// ForceRestart does the chassis power cycle even if the chassis is turned off.
// From the RedFish spec (https://www.dmtf.org/sites/default/files/standards/documents/DSP2046_2018.1.pdf):
//
//	Perform an immediate (non-graceful) shutdown, followed by a restart.
func (i *Ipmi) ForceRestart(ctx context.Context) (status bool, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)

	// Get current power state
	chassisStatus, err := i.client.GetChassisStatus(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get chassis status: %v", err)
	}

	if chassisStatus.PowerIsOn {
		// System is on, do a power cycle
		_, err = i.client.ChassisControl(ctx, ipmi.ChassisControlPowerCycle)
	} else {
		// System is off, just power on
		_, err = i.client.ChassisControl(ctx, ipmi.ChassisControlPowerUp)
	}

	if err != nil {
		return false, fmt.Errorf("chassis control failed: %v", err)
	}
	return true, nil
}

// PowerReset reboots the machine via bmc
func (i *Ipmi) PowerReset(ctx context.Context) (status bool, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	_, err = i.client.ChassisControl(ctx, ipmi.ChassisControlHardReset)
	if err != nil {
		return false, fmt.Errorf("chassis control failed: %v", err)
	}
	return true, nil
}

// PowerCycleBmc reboots the bmc we are connected to
func (i *Ipmi) PowerCycleBmc(ctx context.Context) (status bool, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	err = i.client.ColdReset(ctx)
	if err != nil {
		return false, fmt.Errorf("MC cold reset failed: %v", err)
	}
	return true, nil
}

// PowerResetBmc reboots the bmc we are connected to
func (i *Ipmi) PowerResetBmc(ctx context.Context, resetType string) (ok bool, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	switch strings.ToLower(resetType) {
	case "cold":
		err = i.client.ColdReset(ctx)
	case "warm":
		err = i.client.WarmReset(ctx)
	default:
		return false, fmt.Errorf("unsupported reset type: %s", resetType)
	}
	
	if err != nil {
		return false, fmt.Errorf("MC reset failed: %v", err)
	}
	return true, nil
}

// PowerOn power on the machine via bmc
func (i *Ipmi) PowerOn(ctx context.Context) (status bool, err error) {
	s, err := i.IsOn(ctx)
	if err != nil {
		return false, errors.Wrap(err, "error checking power state")
	}
	if s {
		return true, nil
	}

	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	_, err = i.client.ChassisControl(ctx, ipmi.ChassisControlPowerUp)
	if err != nil {
		return false, fmt.Errorf("chassis control failed: %v", err)
	}
	return true, nil
}

// PowerOnForce power on the machine via bmc even when the machine is already on (Thanks HP!)
func (i *Ipmi) PowerOnForce(ctx context.Context) (status bool, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	_, err = i.client.ChassisControl(ctx, ipmi.ChassisControlPowerUp)
	if err != nil {
		return false, fmt.Errorf("chassis control failed: %v", err)
	}
	return true, nil
}

// PowerOff power off the machine via bmc
func (i *Ipmi) PowerOff(ctx context.Context) (status bool, err error) {
	if on, err := i.IsOn(ctx); err == nil && !on {
		return true, nil
	}
	
	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	_, err = i.client.ChassisControl(ctx, ipmi.ChassisControlPowerDown)
	if err != nil {
		return false, fmt.Errorf("chassis control failed: %v", err)
	}
	return true, nil
}

// PowerSoft power off the machine via bmc
func (i *Ipmi) PowerSoft(ctx context.Context) (status bool, err error) {
	on, _ := i.IsOn(ctx)
	if !on {
		return true, nil
	}

	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	_, err = i.client.ChassisControl(ctx, ipmi.ChassisControlSoftShutdown)
	if err != nil {
		return false, fmt.Errorf("chassis control failed: %v", err)
	}
	return true, nil
}

// PxeOnceEfi makes the machine to boot via pxe once using EFI
func (i *Ipmi) PxeOnceEfi(ctx context.Context) (status bool, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	err = i.client.SetBootDevice(ctx, ipmi.BootDeviceSelectorForcePXE, ipmi.BIOSBootTypeEFI, false)
	if err != nil {
		return false, fmt.Errorf("set boot device failed: %v", err)
	}
	return true, nil
}

// BootDeviceSet sets the next boot device with options
func (i *Ipmi) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	var device ipmi.BootDeviceSelector
	switch strings.ToLower(bootDevice) {
	case "pxe":
		device = ipmi.BootDeviceSelectorForcePXE
	case "disk", "hd":
		device = ipmi.BootDeviceSelectorForceHardDrive
	case "safe":
		device = ipmi.BootDeviceSelectorForceHardDriveSafe
	case "diag":
		device = ipmi.BootDeviceSelectorForceDiagnosticPartition
	case "cdrom", "cd":
		device = ipmi.BootDeviceSelectorForceCDROM
	case "bios", "setup":
		device = ipmi.BootDeviceSelectorForceBIOSSetup
	case "floppy":
		device = ipmi.BootDeviceSelectorForceFloppy
	default:
		device = ipmi.BootDeviceSelectorNoOverride
	}
	
	biosBootType := ipmi.BIOSBootTypeLegacy
	if efiBoot {
		biosBootType = ipmi.BIOSBootTypeEFI
	}
	
	err = i.client.SetBootDevice(ctx, device, biosBootType, setPersistent)
	if err != nil {
		return false, fmt.Errorf("set boot device failed: %v", err)
	}
	return true, nil
}

// PxeOnceMbr makes the machine to boot via pxe once using MBR
func (i *Ipmi) PxeOnceMbr(ctx context.Context) (status bool, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	err = i.client.SetBootDevice(ctx, ipmi.BootDeviceSelectorForcePXE, ipmi.BIOSBootTypeLegacy, false)
	if err != nil {
		return false, fmt.Errorf("set boot device failed: %v", err)
	}
	return true, nil
}

// PxeOnce makes the machine to boot via pxe once using MBR
func (i *Ipmi) PxeOnce(ctx context.Context) (status bool, err error) {
	return i.PxeOnceMbr(ctx)
}

// IsOn tells if a machine is currently powered on
func (i *Ipmi) IsOn(ctx context.Context) (status bool, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return false, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	chassisStatus, err := i.client.GetChassisStatus(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get chassis status: %v", err)
	}
	return chassisStatus.PowerIsOn, nil
}

// PowerState returns the current power state of the machine
func (i *Ipmi) PowerState(ctx context.Context) (state string, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return "", fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	chassisStatus, err := i.client.GetChassisStatus(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get chassis status: %v", err)
	}
	
	if chassisStatus.PowerIsOn {
		return "Chassis Power is on", nil
	}
	return "Chassis Power is off", nil
}

// ReadUsers list all BMC users
func (i *Ipmi) ReadUsers(ctx context.Context) (users []map[string]string, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	// Try to get user information for user IDs 1-16 (typical range)
	// Since GetUsers might not be available, we'll iterate through user IDs
	for userID := uint8(1); userID <= 16; userID++ {
		userAccess, err := i.client.GetUserAccess(ctx, 1, userID)
		if err != nil {
			// Skip users that don't exist or can't be accessed
			continue
		}
		
		// Get username for this user ID
		userNameResp, err := i.client.GetUsername(ctx, userID)
		if err != nil {
			// Skip users that can't be queried
			continue
		}
		
		if userNameResp.Username == "" {
			// Skip users without names
			continue
		}
		
		users = append(users, map[string]string{
			"ID":               fmt.Sprintf("%d", userID),
			"Name":             userNameResp.Username,
			"Callin":           fmt.Sprintf("%t", userAccess.CallbackOnly),
			"Link Auth":        fmt.Sprintf("%t", userAccess.LinkAuthEnabled),
			"IPMI Msg":         fmt.Sprintf("%t", userAccess.IPMIMessagingEnabled),
			"Channel Priv Limit": fmt.Sprintf("%v", userAccess.MaxPrivLevel),
		})
	}
	
	return users, nil
}

// ClearSystemEventLog clears the system event log
func (i *Ipmi) ClearSystemEventLog(ctx context.Context) (err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	// Use 0x4321 as the clear operation code (standard IPMI clear operation)
	_, err = i.client.ClearSEL(ctx, 0x4321)
	if err != nil {
		return fmt.Errorf("failed to clear SEL: %v", err)
	}
	return nil
}

// GetSystemEventLog returns the system event log entries in ID, Timestamp, Description, Message format
func (i *Ipmi) GetSystemEventLog(ctx context.Context) (entries [][]string, err error) {
	raw, err := i.GetSystemEventLogRaw(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting system event log")
	}

	entries = parseSystemEventLog(raw)
	return entries, nil
}

// GetSystemEventLogRaw returns the raw SEL output
func (i *Ipmi) GetSystemEventLogRaw(ctx context.Context) (eventlog string, err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return "", fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	// Get all SEL entries starting from record ID 0
	selEntries, err := i.client.GetSELEntries(ctx, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get SEL entries: %v", err)
	}
	
	// Format SEL entries into the expected raw format for compatibility
	var lines []string
	lines = append(lines, "   SEL Record ID          | Date/Time         | Sensor Name      | Event Dir  | Event Data")
	
	for _, entry := range selEntries {
		if entry.Standard != nil {
			timestamp := entry.Standard.Timestamp.Format("01/02/2006 | 15:04:05")
			sensorName := fmt.Sprintf("Sensor %d", entry.Standard.SensorNumber)
			eventDir := "Asserted"
			if entry.Standard.EventDir == ipmi.EventDirDeassertion {
				eventDir = "Deasserted"
			}
			eventData := fmt.Sprintf("0x%02x 0x%02x 0x%02x", 
				entry.Standard.EventData.EventData1,
				entry.Standard.EventData.EventData2,
				entry.Standard.EventData.EventData3)
			
			line := fmt.Sprintf(" %04x | %s | %-16s | %-10s | %s",
				entry.RecordID, timestamp, sensorName, eventDir, eventData)
			lines = append(lines, line)
		}
	}
	
	return strings.Join(lines, "\n"), nil
}

func (i *Ipmi) DeactivateSOL(ctx context.Context) (err error) {
	if err := i.ensureConnected(ctx); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)

	_, err = i.client.DeactivatePayload(ctx, &ipmi.DeactivatePayloadRequest{
		PayloadType:     ipmi.PayloadTypeSOL,
		PayloadInstance: 0,
	})
	if err != nil {
		// 0x80 means SOL was already deactivated; treat as success.
		var respErr *ipmi.ResponseError
		if errors.As(err, &respErr) && respErr.CompletionCode() == 0x80 {
			return nil
		}
		return fmt.Errorf("failed to deactivate SOL payload: %w", err)
	}
	return nil
}

// SendPowerDiag tells the BMC to issue an NMI to the device
func (i *Ipmi) SendPowerDiag(ctx context.Context) error {
	if err := i.ensureConnected(ctx); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer i.client.Close(ctx)
	
	_, err := i.client.ChassisControl(ctx, ipmi.ChassisControlDiagnosticInterrupt)
	if err != nil {
		return errors.Wrap(err, "failed sending power diag")
	}
	return nil
}
