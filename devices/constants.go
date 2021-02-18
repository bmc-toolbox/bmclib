package devices

const (
	// Unknown is the constant that defines unknown things
	Unknown = "Unknown"

	// Vendor constants

	// HP is the constant that defines the vendor HP
	HP = "HP"
	// Dell is the constant that defines the vendor Dell
	Dell = "Dell"
	// Supermicro is the constant that defines the vendor Supermicro
	Supermicro = "Supermicro"
	// Cloudline is the constant that defines the cloudlines
	Cloudline = "Cloudline"
	// Quanta is the contant to identify Quanta hardware
	Quanta = "Quanta"
	// Common is the constant of thinks we could use across multiple vendors
	Common = "Common"

	// Power Constants

	// Grid describes the power redundancy mode when using grid redundancy
	Grid = "Grid"

	// PowerSupply describes the power redundancy mode when using power supply redundancy
	PowerSupply = "PowerSupply"

	// NoRedundancy describes the power redundancy mode we don't have redundancy
	NoRedundancy = "NoRedundancy"

	// Hardware constants

	// BladeHwType is the constant defining the blade hw type
	BladeHwType = "blade"
	// DiscreteHwType is the constant defining the Discrete hw type
	DiscreteHwType = "discrete"
	// ChassisHwType is the constant defining the chassis hw type
	ChassisHwType = "chassis"
)

// ListSupportedVendors  returns a list of supported vendors
func ListSupportedVendors() []string {
	return []string{HP, Dell, Supermicro}
}
