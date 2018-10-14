#### Templating support
The configuration for each asset is rendered from [configuration.yml](../master/cfg/configuration.yml)
templating is supported for this configuration file using the pretty cool [Plush](https://github.com/gobuffalo/plush/)


Variables for the template are made available based on asset attributes,
which were retrieved either from the enc/inventory lookup tool or from the BMC itself.

Note: All string variables are downcased.

#### Supported variables

Template variable   |    type   | info                 |
:-----------------  | :-------: | :------------------: |
vendor              |   string  | Asset vendor (HP/Dell/Supermicro |
location            |   string  | Asset location - a datacenter identifier, e.g ams5 |
assetType           |   string  | Type of asset (server/chassis)                     |
model               |   string  | Model number of the asset (idrac8/ilo5/m1000e)     |
serial              |   string  | Serial/Identification number for the asset.        |
ipaddress           |   string  | IP Address of the asset (if its a chassis with multiple IPs this is the active IP)                  |
extra["state"]      |   string  | Extra attribute from the ENC, 'state' identifies the inventory state of the asset (live/needs-setup)|
extra["company"]    |   string  | Extra attribute from the ENC, 'company' identifies the owner of the asset.                          |

Any new variables exposed that might be metadata specific to a company or some business specific logic, 
should end up in 'extra' (map[string]string).

#### Examples

Conditional declaration of configuration resources.
```
# Delare license configuration resource if vendor is 'hp'
<%= if ( vendor == "hp" ) { %>
  license:
    key: FOOOBARR5432
<% } %>


#Based on various asset attributes, declare SetupChassis config resource.
<% if ( assetType == "chassis" &&
        extra["state"] == "needs-setup" &&
        extra["company"] == "skynet" ) { %>
setupChassis:
  ipmiOverLan:
    enable: true
  flexAddress:
    enable: false
  dynamicPower:
    enable: false
<% } %>
```

General variable interpolation.
```
ldapGroup:
  - role: admin
    group: cn=<%= vendor %>,cn=bmcAdmins
    groupBaseDn: ou=Group,dc=example,dc=com #the baseDn to lookup group in.
    enable: true
  - role: user
    group: cn=<%= vendor%>,cn=bmcUsers #the group in ldap
    groupBaseDn: ou=Group,dc=example,dc=com #the baseDn to lookup group in.
```




