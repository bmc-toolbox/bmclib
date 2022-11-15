module github.com/bmc-toolbox/bmclib/v2

go 1.18

require (
	github.com/ammmze/go-amt v0.0.4
	github.com/bmc-toolbox/bmclib v0.5.4
	github.com/bmc-toolbox/common v0.0.0-20220707135204-5368ecd5d175
	github.com/bombsimon/logrusr/v2 v2.0.1
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/stdr v1.2.2
	github.com/google/go-cmp v0.5.8
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jacobweinstock/registrar v0.4.6
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stmcginnis/gofish v0.13.1-0.20221107140645-5cc43fad050f
	github.com/stretchr/testify v1.7.2
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
	golang.org/x/net v0.0.0-20220615171555-694bf12d69de
	gopkg.in/go-playground/assert.v1 v1.2.1
)

require (
	github.com/VictorLowther/simplexml v0.0.0-20180716164440-0bff93621230 // indirect
	github.com/VictorLowther/soap v0.0.0-20150314151524-8e36fca84b22 // indirect
	github.com/ammmze/wsman v0.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.5.1 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/tools v0.1.8 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/ammmze/go-amt => github.com/jacobweinstock/go-amt v0.0.0-20221115190754-36c39f21b864

replace github.com/ammmze/wsman => github.com/jacobweinstock/wsman v0.0.0-20221115191137-e06ce6a5341d
