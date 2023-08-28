package utils

var RegisrtyNameToMintscanName = map[string]string{
	"cosmos":        "cosmos",
	"cosmoshub":     "cosmos",
	"osmosis":       "osmosis",
	"regen":         "regen",
	"akash":         "akash",
	"stride":        "stride",
	"juno":          "juno",
	"umee":          "umee",
	"omniflixhub":   "omniflix",
	"axelar":        "axelar",
	"bandchain":     "bandchain",
	"comdex":        "comdex",
	"desmos":        "desmos",
	"emoney":        "emoney",
	"evmos":         "evmos",
	"gravitybridge": "gravity-bridge",
	"tgrade":        "tgrade",
	"stargaze":      "stargaze",
	"sentinel":      "sentinel",
	"quicksilver":   "quicksilver",
	"persistence":   "persistence",
}

type DenomInfo struct {
	BaseDenom    string
	DisplayDenom string
	DenomUnits   int64
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
		BaseDenom:    "uatom",
		DisplayDenom: "ATOM",
		DenomUnits:   6,
	},
	"cosmoshub": {
		BaseDenom:    "uatom",
		DisplayDenom: "ATOM",
		DenomUnits:   6,
	},
	"osmosis": {
		BaseDenom:    "uosmo",
		DisplayDenom: "OSMO",
		DenomUnits:   6,
	},
	"regen": {
		BaseDenom:    "uregen",
		DisplayDenom: "REGEN",
		DenomUnits:   6,
	},
	"akash": {
		BaseDenom:    "uakt",
		DisplayDenom: "AKT",
		DenomUnits:   6,
	},
	"stride": {
		BaseDenom:    "ustride",
		DisplayDenom: "STRIDE",
		DenomUnits:   6,
	},
	"juno": {
		BaseDenom:    "ujuno",
		DisplayDenom: "JUNO",
		DenomUnits:   6,
	},
	"umee": {
		BaseDenom:    "uumee",
		DisplayDenom: "UMEE",
		DenomUnits:   6,
	},
	"omniflixhub": {
		BaseDenom:    "uflix",
		DisplayDenom: "FLIX",
		DenomUnits:   6,
	},
	"axelar": {
		BaseDenom:    "uaxl",
		DisplayDenom: "AXL",
		DenomUnits:   6,
	},
	"bandchain": {
		BaseDenom:    "uband",
		DisplayDenom: "BAND",
		DenomUnits:   6,
	},
	"comdex": {
		BaseDenom:    "ucmdx",
		DisplayDenom: "CMDX",
		DenomUnits:   6,
	},
	"desmos": {
		BaseDenom:    "udsm",
		DisplayDenom: "DSM",

		DenomUnits: 6,
	},
	"emoney": {
		BaseDenom:    "ungm",
		DisplayDenom: "NGM",
		DenomUnits:   6,
	},
	"evmos": {
		BaseDenom:    "aevmos",
		DisplayDenom: "EVMOS",
		DenomUnits:   18,
	},
	"gravitybridge": {
		BaseDenom:    "ugraviton",
		DisplayDenom: "GRAV",
		DenomUnits:   6,
	},
	"tgrade": {
		BaseDenom:    "utgd",
		DisplayDenom: "TGD",
		DenomUnits:   6,
	},
	"stargaze": {
		BaseDenom:    "ustars",
		DisplayDenom: "STARS",
		DenomUnits:   6,
	},
	"sentinel": {
		BaseDenom:    "udvpn",
		DisplayDenom: "DVPN",
		DenomUnits:   6,
	},
	"quicksilver": {
		BaseDenom:    "uqck",
		DisplayDenom: "QCK",
		DenomUnits:   6,
	},
	"persistence": {
		BaseDenom:    "uxprt",
		DisplayDenom: "XPRT",
		DenomUnits:   6,
	},
	"passage": {
		BaseDenom:    "upasg",
		DisplayDenom: "PSG",
		DenomUnits:   6,
	},
}
