package utils

type DenomInfo struct {
	BaseDenom  string
	DenomUnits int64
}

var GovV1Support = map[string]map[string]bool{
	"cosmos": {
		"govv1_enabled": false,
		"authz_enabled": true,
	},
	"osmosis": {
		"govv1_enabled": true,
		"authz_enabled": true,
	},
	"regen": {
		"govv1_enabled": true,
		"authz_enabled": true,
	},
	"akash": {
		"govv1_enabled": false,
		"authz_enabled": true,
	},
	"juno": {
		"govv1_enabled": false,
		"authz_enabled": true,
	},
	"umee": {
		"govv1_enabled": true,
		"authz_enabled": true,
	},
	"omniflix": {
		"govv1_enabled": false,
		"authz_enabled": true,
	},
	"comdex": {
		"govv1_enabled": false,
		"authz_enabled": true,
	},
	"mars": {
		"govv1_enabled": true,
		"authz_enabled": true,
	},
	"desmos": {
		"govv1_enabled": false,
		"authz_enabled": true,
	},
	"evmos": {
		"govv1_enabled": true,
		"authz_enabled": true,
	},
	"stargaze": {
		"govv1_enabled": false,
		"authz_enabled": true,
	},
	"quicksilver": {
		"govv1_enabled": true,
		"authz_enabled": true,
	},
	"crescent": {
		"govv1_enabled": false,
		"authz_enabled": true,
	},
	"passage": {
		"govv1_enabled": false,
		"authz_enabled": true,
	},
}

var ChainNameToDenomInfo = map[string]DenomInfo{
	"cosmos": {
		BaseDenom:  "uatom",
		DenomUnits: 6,
	},
	"cosmoshub": {
		BaseDenom:  "uatom",
		DenomUnits: 6,
	},
	"osmosis": {
		BaseDenom:  "uosmo",
		DenomUnits: 6,
	},
	"regen": {
		BaseDenom:  "uregen",
		DenomUnits: 6,
	},
	"akash": {
		BaseDenom:  "uakt",
		DenomUnits: 6,
	},
	"stride": {
		BaseDenom:  "ustride",
		DenomUnits: 6,
	},
	"juno": {
		BaseDenom:  "ujuno",
		DenomUnits: 6,
	},
	"umee": {
		BaseDenom:  "uumee",
		DenomUnits: 6,
	},
	"omniflixhub": {
		BaseDenom:  "uflix",
		DenomUnits: 6,
	},
	"axelar": {
		BaseDenom:  "uaxl",
		DenomUnits: 6,
	},
	"bandchain": {
		BaseDenom:  "uband",
		DenomUnits: 6,
	},
	"comdex": {
		BaseDenom:  "ucmdx",
		DenomUnits: 6,
	},
	"desmos": {
		BaseDenom:  "udsm",
		DenomUnits: 6,
	},
	"emoney": {
		BaseDenom:  "ungm",
		DenomUnits: 6,
	},
	"evmos": {
		BaseDenom:  "aevmos",
		DenomUnits: 18,
	},
	"gravitybridge": {
		BaseDenom:  "ugraviton",
		DenomUnits: 6,
	},
	"tgrade": {
		BaseDenom:  "utgd",
		DenomUnits: 6,
	},
	"stargaze": {
		BaseDenom:  "ustars",
		DenomUnits: 6,
	},
	"sentinel": {
		BaseDenom:  "udvpn",
		DenomUnits: 6,
	},
	"quicksilver": {
		BaseDenom:  "uqck",
		DenomUnits: 6,
	},
	"persistence": {
		BaseDenom:  "uxprt",
		DenomUnits: 6,
	},
	"passage": {
		BaseDenom:  "upasg",
		DenomUnits: 6,
	},
}
