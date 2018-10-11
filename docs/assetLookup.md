### About

This document attempts to explain the design and behaviour of the asset lookup binary,
which can be deployed along with bmcbutler to query assets and their attributes
from the asset inventory.

The idea is similar to a ENC (external node classifier),
in this case given a server/chassis serial/ip its attributes are returned,
the executable could also lookup the inventory given a limit and offset.

### Why

Inventory lookup being handled by a binary of a users choice,
allows the business logic to maintained separately and decoupled from bmcbutler.

### How

The executable needs to support querying the asset database
for either the list of assets to configure along with its attributes,
or given a list of IP(s)/Serial(s) its attributes are returned.

The executable supports the following args two args,

enc, inventory

#### Output format

Except JSON, nothing else should be dumped to stdout by the lookup tool,
the executable when run dumps the asset attributes retrieved in a JSON
format.

The minimum requirement for the JSON output of the executable is to have the "data" field.
Each asset is returned with its identifier (a serial or an ID) as the key and the asset attributes
as the values, see examples below.

#### Exit codes

The executable must exit with a non zero code if the lookup fails for whatever reason,
the example below lists a few possible exit codes.

```
ErrConfig         = 132 //Error loading config
ErrInvalidArgs    = 22  //Not enough args given
ErrInvResponse    = 133 //Error in inventory lookup response
ErrInvNoResults   = 134 //No results to return
ErrJsonMarshal    = 135 //Unable to marshal data to JSON
ErrEndOfInvAssets = 136 //Reached the end of assets in inventory
```
#### Asset attribute lookup

Lookup assets by given IP(s) address

```
$ assetlookup enc --ips 10.193.251.111,10.193.251.112 | jq
{
  "data": {
    "4A1231F92": {
      "location": "ams6",
      "ipaddress": [
        "10.193.251.111"
      ],
      "extras": {
        "status": "claimed",
        "company": "Booking.com",
        "assetType": "",
        "live_assets": null
      }
    },
    "C6123FA": {
      "location": "ams5",
      "ipaddress": [
        "10.193.251.112"
      ],
      "extras": {
        "status": "needs-setup",
        "company": "Booking.com",
        "assetType": "",
        "live_assets": null
      }
    }
  }
}
```
Lookup assets by given serial(s)

$ assetlookup enc --serials 4NYYR12  | jq
{
  "data": {
    "4NYYR12": {
      "location": "ams6",
      "ipaddress": [
        "10.193.251.111"
      ],
      "extras": {
        "status": "claimed",
        "company": "Booking.com",
        "assetType": "",
        "live_assets": null
      }
    }
  }
}


##### Inventory assets listing

Lookup a server at the given location.

```
$ assetlookup inventory --server --limit 1 --offset 10 --location lhr4 | jq
{
  "data": {
    "SERIAL12312": {
      "location": "lhr4",
      "ipaddress": [
        "10.183.203.142"
      ],
      "extras": {
        "status": "live",
        "company": "",
        "assetType": "server",
        "live_assets": null
      }
    }
  },
  "offset": 10,
  "limit": 1
}
```

Lookup Chassis asset.
```
$ assetlookup inventory --chassis --limit 1 --offset 10 --location lhr4 | jq
{
  "data": {
    "SERI47": {
      "location": "lhr4",
      "ipaddress": [
        "10.183.185.118",
        "10.183.185.101"
      ],
      "extras": {
        "status": "installed",
        "company": "Booking.com",
        "assetType": "chassis",
        "live_assets": [
          "CZ3ASDAA",
          "C5359RE3P",
          "BASDRE3S"
        ]
      }
    }
  },
  "offset": 10,
  "limit": 1
}
```
