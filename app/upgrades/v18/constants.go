package v18

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	UpgradeName = "v18"

	// Redemption rate bounds updated to give ~3 months of slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdk.MustNewDecFromStr("0.05")
	RedemptionRateOuterMaxAdjustment = sdk.MustNewDecFromStr("0.10")

	// Osmosis will have a slighly larger buffer with the redemption rate
	// since their yield is less predictable
	OsmosisChainId              = "osmosis-1"
	OsmosisRedemptionRateBuffer = sdk.MustNewDecFromStr("0.02")

	// Terra chain ID for delegation changes in progress
	TerraChainId = "phoenix-1"

	// Get Initial Redemption Rates for Unbonding Records Migration
	RedemptionRatesAtTimeOfProp = map[string]sdk.Dec{
		"comdex-1":     sdk.MustNewDecFromStr("1.204883527965105396"),
		"cosmoshub-4":  sdk.MustNewDecFromStr("1.299886984330871277"),
		"evmos_9001-2": sdk.MustNewDecFromStr("1.492732862363044751"),
		"injective-1":  sdk.MustNewDecFromStr("1.216027814303310584"),
		"juno-1":       sdk.MustNewDecFromStr("1.418690442281976982"),
		"osmosis-1":    sdk.MustNewDecFromStr("1.201662502920632779"),
		"phoenix-1":    sdk.MustNewDecFromStr("1.178584742254853106"),
		"sommelier-3":  sdk.MustNewDecFromStr("1.025900897761638723"),
		"stargaze-1":   sdk.MustNewDecFromStr("1.430486928659223287"),
		"umee-1":       sdk.MustNewDecFromStr("1.128892781103330908"),
	}

	// Get Amount Unbonded for each HostZone for Unbonding Records Migration
	StartingEstimateEpoch     = uint64(508)
	RedemptionRatesBeforeProp = map[string]map[uint64]sdk.Dec{
		"juno-1": {
			495: sdk.MustNewDecFromStr("1.412164551270598"),
			496: sdk.MustNewDecFromStr("1.412164551270598"),
			497: sdk.MustNewDecFromStr("1.412164551270598"),
			500: sdk.MustNewDecFromStr("1.4161495546072012"),
			501: sdk.MustNewDecFromStr("1.4161495546072012"),
			503: sdk.MustNewDecFromStr("1.4161495546072012"),
			504: sdk.MustNewDecFromStr("1.4161495546072012"),
		},
		"phoenix-1": {
			496: sdk.MustNewDecFromStr("1.1740619020285001"),
			498: sdk.MustNewDecFromStr("1.1740619020285001"),
			499: sdk.MustNewDecFromStr("1.1740619020285001"),
			500: sdk.MustNewDecFromStr("1.1757224643748854"),
			503: sdk.MustNewDecFromStr("1.1757224643748854"),
			504: sdk.MustNewDecFromStr("1.176553937681711"),
			505: sdk.MustNewDecFromStr("1.176553937681711"),
			506: sdk.MustNewDecFromStr("1.176553937681711"),
			507: sdk.MustNewDecFromStr("1.176553937681711"),
		},
		"sommelier-3": {
			495: sdk.MustNewDecFromStr("1.0241481197817144"),
			496: sdk.MustNewDecFromStr("1.0241481197817144"),
			497: sdk.MustNewDecFromStr("1.0241481197817144"),
			499: sdk.MustNewDecFromStr("1.0241481197817144"),
			501: sdk.MustNewDecFromStr("1.025236900070852"),
			502: sdk.MustNewDecFromStr("1.025236900070852"),
			503: sdk.MustNewDecFromStr("1.025236900070852"),
			504: sdk.MustNewDecFromStr("1.025236900070852"),
		},
		"cosmoshub-4": {
			496: sdk.MustNewDecFromStr("1.2938404518607025"),
			497: sdk.MustNewDecFromStr("1.2938404518607025"),
			498: sdk.MustNewDecFromStr("1.2938404518607025"),
			499: sdk.MustNewDecFromStr("1.2938404518607025"),
			500: sdk.MustNewDecFromStr("1.2957672912922817"),
			501: sdk.MustNewDecFromStr("1.2957672912922817"),
			502: sdk.MustNewDecFromStr("1.2957672912922817"),
			503: sdk.MustNewDecFromStr("1.2957672912922817"),
			504: sdk.MustNewDecFromStr("1.296926394723948"),
			505: sdk.MustNewDecFromStr("1.296926394723948"),
			506: sdk.MustNewDecFromStr("1.296926394723948"),
			507: sdk.MustNewDecFromStr("1.296926394723948"),
		},
		"comdex-1": {
			496: sdk.MustNewDecFromStr("1.1963306878344375"),
			497: sdk.MustNewDecFromStr("1.1963306878344375"),
			498: sdk.MustNewDecFromStr("1.1963306878344375"),
			499: sdk.MustNewDecFromStr("1.1963306878344375"),
			500: sdk.MustNewDecFromStr("1.1994537074221134"),
			501: sdk.MustNewDecFromStr("1.1994537074221134"),
			502: sdk.MustNewDecFromStr("1.1994537074221134"),
			503: sdk.MustNewDecFromStr("1.1994537074221134"),
			504: sdk.MustNewDecFromStr("1.2019746297343605"),
			505: sdk.MustNewDecFromStr("1.2019746297343605"),
			506: sdk.MustNewDecFromStr("1.2019746297343605"),
			507: sdk.MustNewDecFromStr("1.2019746297343605"),
		},
		"injective-1": {
			464: sdk.MustNewDecFromStr("1.10904028152176"),
			465: sdk.MustNewDecFromStr("1.1092232046811195"),
			466: sdk.MustNewDecFromStr("1.1094104738505122"),
			467: sdk.MustNewDecFromStr("1.109660102119856"),
			468: sdk.MustNewDecFromStr("1.1099206471560683"),
			469: sdk.MustNewDecFromStr("1.1101781888690843"),
			470: sdk.MustNewDecFromStr("1.1104928343163862"),
			471: sdk.MustNewDecFromStr("1.1106814727683936"),
			472: sdk.MustNewDecFromStr("1.1109147705303473"),
			473: sdk.MustNewDecFromStr("1.1111483631454906"),
			474: sdk.MustNewDecFromStr("1.1113789833325327"),
			475: sdk.MustNewDecFromStr("1.1115865207841595"),
			476: sdk.MustNewDecFromStr("1.1118256565192843"),
			477: sdk.MustNewDecFromStr("1.112062977242558"),
			478: sdk.MustNewDecFromStr("1.112305089405149"),
			479: sdk.MustNewDecFromStr("1.1125496812740654"),
			480: sdk.MustNewDecFromStr("1.112796928321449"),
			481: sdk.MustNewDecFromStr("1.113045979582398"),
			482: sdk.MustNewDecFromStr("1.1133578645679472"),
			483: sdk.MustNewDecFromStr("1.1135463131500978"),
			484: sdk.MustNewDecFromStr("1.113862639530537"),
			485: sdk.MustNewDecFromStr("1.1140510045259582"),
			486: sdk.MustNewDecFromStr("1.114295573398525"),
			487: sdk.MustNewDecFromStr("1.1145990588175787"),
			488: sdk.MustNewDecFromStr("1.114779498371232"),
			489: sdk.MustNewDecFromStr("1.1150839991290917"),
			498: sdk.MustNewDecFromStr("1.1170896901082266"),
			499: sdk.MustNewDecFromStr("1.1498981693771557"),
			500: sdk.MustNewDecFromStr("1.209508137205966"),
			501: sdk.MustNewDecFromStr("1.209985009275008"),
			502: sdk.MustNewDecFromStr("1.210478332327813"),
			503: sdk.MustNewDecFromStr("1.2109676716098068"),
			504: sdk.MustNewDecFromStr("1.2130924701151315"),
			505: sdk.MustNewDecFromStr("1.2136053525521355"),
			507: sdk.MustNewDecFromStr("1.21455566769327"),
		},
		"evmos_9001-2": {
			499: sdk.MustNewDecFromStr("1.4895991845634247"),
			500: sdk.MustNewDecFromStr("1.4895991845634247"),
			501: sdk.MustNewDecFromStr("1.490098715761824"),
			502: sdk.MustNewDecFromStr("1.490098715761824"),
			503: sdk.MustNewDecFromStr("1.490098715761824"),
			504: sdk.MustNewDecFromStr("1.4910458236916064"),
			505: sdk.MustNewDecFromStr("1.4910458236916064"),
		},
		"osmosis-1": {
			498: sdk.MustNewDecFromStr("1.1984190041836773"),
			499: sdk.MustNewDecFromStr("1.1984190041836773"),
			500: sdk.MustNewDecFromStr("1.1984190041836773"),
			501: sdk.MustNewDecFromStr("1.1991174772238702"),
			502: sdk.MustNewDecFromStr("1.1991174772238702"),
			503: sdk.MustNewDecFromStr("1.1991174772238702"),
			504: sdk.MustNewDecFromStr("1.2003177583397713"),
			505: sdk.MustNewDecFromStr("1.2003177583397713"),
			506: sdk.MustNewDecFromStr("1.2003177583397713"),
		},
		"stargaze-1": {
			498: sdk.MustNewDecFromStr("1.4246347073913794"),
			499: sdk.MustNewDecFromStr("1.4246347073913794"),
			500: sdk.MustNewDecFromStr("1.4246347073913794"),
			501: sdk.MustNewDecFromStr("1.4267297754925006"),
			502: sdk.MustNewDecFromStr("1.4267297754925006"),
			503: sdk.MustNewDecFromStr("1.4267297754925006"),
			504: sdk.MustNewDecFromStr("1.4279528400269015"),
			505: sdk.MustNewDecFromStr("1.4279528400269015"),
			506: sdk.MustNewDecFromStr("1.4279528400269015"),
		},
		"umee-1": {
			505: sdk.MustNewDecFromStr("1.1266406527137283"),
		},
	}
)
