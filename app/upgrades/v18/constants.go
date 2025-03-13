package v18

import (
	sdkmath "cosmossdk.io/math"
)

var (
	UpgradeName = "v18"

	// Redemption rate bounds updated to give ~3 months of slack on outer bounds
	RedemptionRateOuterMinAdjustment = sdkmath.LegacyMustNewDecFromStr("0.05")
	RedemptionRateOuterMaxAdjustment = sdkmath.LegacyMustNewDecFromStr("0.10")

	// Osmosis will have a slighly larger buffer with the redemption rate
	// since their yield is less predictable
	OsmosisChainId              = "osmosis-1"
	OsmosisRedemptionRateBuffer = sdkmath.LegacyMustNewDecFromStr("0.02")

	// Terra chain ID for delegation changes in progress
	TerraChainId = "phoenix-1"

	// Prop 228 info
	Strd                       = "ustrd"
	Prop228ProposalId          = uint64(228)
	Prop228SendAmount          = sdkmath.NewInt(9_000_000_000_000)
	IncentiveProgramAddress    = "stride1tlxk4as9sgpqkh42cfaxqja0mdj6qculqshy0gg3glazmrnx3y8s8gsvqk"
	StrideFoundationAddress_F4 = "stride1yz3mp7c2m739nftfrv5r3h6j64aqp95f3degpf"

	// Get Initial Redemption Rates for Unbonding Records Migration
	RedemptionRatesAtTimeOfProp = map[string]sdkmath.LegacyDec{
		"comdex-1":     sdkmath.LegacyMustNewDecFromStr("1.204883527965105396"),
		"cosmoshub-4":  sdkmath.LegacyMustNewDecFromStr("1.299886984330871277"),
		"evmos_9001-2": sdkmath.LegacyMustNewDecFromStr("1.492732862363044751"),
		"injective-1":  sdkmath.LegacyMustNewDecFromStr("1.216027814303310584"),
		"juno-1":       sdkmath.LegacyMustNewDecFromStr("1.418690442281976982"),
		"osmosis-1":    sdkmath.LegacyMustNewDecFromStr("1.201662502920632779"),
		"phoenix-1":    sdkmath.LegacyMustNewDecFromStr("1.178584742254853106"),
		"sommelier-3":  sdkmath.LegacyMustNewDecFromStr("1.025900897761638723"),
		"stargaze-1":   sdkmath.LegacyMustNewDecFromStr("1.430486928659223287"),
		"umee-1":       sdkmath.LegacyMustNewDecFromStr("1.128892781103330908"),
	}

	// Get Amount Unbonded for each HostZone for Unbonding Records Migration
	StartingEstimateEpoch     = uint64(508)
	RedemptionRatesBeforeProp = map[string]map[uint64]sdkmath.LegacyDec{
		"juno-1": {
			495: sdkmath.LegacyMustNewDecFromStr("1.412164551270598"),
			496: sdkmath.LegacyMustNewDecFromStr("1.412164551270598"),
			497: sdkmath.LegacyMustNewDecFromStr("1.412164551270598"),
			500: sdkmath.LegacyMustNewDecFromStr("1.4161495546072012"),
			501: sdkmath.LegacyMustNewDecFromStr("1.4161495546072012"),
			503: sdkmath.LegacyMustNewDecFromStr("1.4161495546072012"),
			504: sdkmath.LegacyMustNewDecFromStr("1.4161495546072012"),
			505: sdkmath.LegacyMustNewDecFromStr("1.417724248601981"),
			507: sdkmath.LegacyMustNewDecFromStr("1.417724248601981"),
			508: sdkmath.LegacyMustNewDecFromStr("1.417724248601981"),
		},
		"phoenix-1": {
			496: sdkmath.LegacyMustNewDecFromStr("1.1740619020285001"),
			498: sdkmath.LegacyMustNewDecFromStr("1.1740619020285001"),
			499: sdkmath.LegacyMustNewDecFromStr("1.1740619020285001"),
			500: sdkmath.LegacyMustNewDecFromStr("1.1757224643748854"),
			503: sdkmath.LegacyMustNewDecFromStr("1.1757224643748854"),
			504: sdkmath.LegacyMustNewDecFromStr("1.176553937681711"),
			505: sdkmath.LegacyMustNewDecFromStr("1.176553937681711"),
			506: sdkmath.LegacyMustNewDecFromStr("1.176553937681711"),
			507: sdkmath.LegacyMustNewDecFromStr("1.176553937681711"),
		},
		"sommelier-3": {
			495: sdkmath.LegacyMustNewDecFromStr("1.0241481197817144"),
			496: sdkmath.LegacyMustNewDecFromStr("1.0241481197817144"),
			497: sdkmath.LegacyMustNewDecFromStr("1.0241481197817144"),
			499: sdkmath.LegacyMustNewDecFromStr("1.0241481197817144"),
			501: sdkmath.LegacyMustNewDecFromStr("1.025236900070852"),
			502: sdkmath.LegacyMustNewDecFromStr("1.025236900070852"),
			503: sdkmath.LegacyMustNewDecFromStr("1.025236900070852"),
			504: sdkmath.LegacyMustNewDecFromStr("1.025236900070852"),
			505: sdkmath.LegacyMustNewDecFromStr("1.0259008616651284"),
			507: sdkmath.LegacyMustNewDecFromStr("1.0259008616651284"),
			508: sdkmath.LegacyMustNewDecFromStr("1.0259008616651284"),
			509: sdkmath.LegacyMustNewDecFromStr("1.0259008616651284"),
		},
		"cosmoshub-4": {
			496: sdkmath.LegacyMustNewDecFromStr("1.2938404518607025"),
			497: sdkmath.LegacyMustNewDecFromStr("1.2938404518607025"),
			498: sdkmath.LegacyMustNewDecFromStr("1.2938404518607025"),
			499: sdkmath.LegacyMustNewDecFromStr("1.2938404518607025"),
			500: sdkmath.LegacyMustNewDecFromStr("1.2957672912922817"),
			501: sdkmath.LegacyMustNewDecFromStr("1.2957672912922817"),
			502: sdkmath.LegacyMustNewDecFromStr("1.2957672912922817"),
			503: sdkmath.LegacyMustNewDecFromStr("1.2957672912922817"),
			504: sdkmath.LegacyMustNewDecFromStr("1.296926394723948"),
			505: sdkmath.LegacyMustNewDecFromStr("1.296926394723948"),
			506: sdkmath.LegacyMustNewDecFromStr("1.296926394723948"),
			507: sdkmath.LegacyMustNewDecFromStr("1.296926394723948"),
		},
		"comdex-1": {
			496: sdkmath.LegacyMustNewDecFromStr("1.1963306878344375"),
			497: sdkmath.LegacyMustNewDecFromStr("1.1963306878344375"),
			498: sdkmath.LegacyMustNewDecFromStr("1.1963306878344375"),
			499: sdkmath.LegacyMustNewDecFromStr("1.1963306878344375"),
			500: sdkmath.LegacyMustNewDecFromStr("1.1994537074221134"),
			501: sdkmath.LegacyMustNewDecFromStr("1.1994537074221134"),
			502: sdkmath.LegacyMustNewDecFromStr("1.1994537074221134"),
			503: sdkmath.LegacyMustNewDecFromStr("1.1994537074221134"),
			504: sdkmath.LegacyMustNewDecFromStr("1.2019746297343605"),
			505: sdkmath.LegacyMustNewDecFromStr("1.2019746297343605"),
			506: sdkmath.LegacyMustNewDecFromStr("1.2019746297343605"),
			507: sdkmath.LegacyMustNewDecFromStr("1.2019746297343605"),
		},
		"injective-1": {
			464: sdkmath.LegacyMustNewDecFromStr("1.10904028152176"),
			465: sdkmath.LegacyMustNewDecFromStr("1.1092232046811195"),
			466: sdkmath.LegacyMustNewDecFromStr("1.1094104738505122"),
			467: sdkmath.LegacyMustNewDecFromStr("1.109660102119856"),
			468: sdkmath.LegacyMustNewDecFromStr("1.1099206471560683"),
			469: sdkmath.LegacyMustNewDecFromStr("1.1101781888690843"),
			470: sdkmath.LegacyMustNewDecFromStr("1.1104928343163862"),
			471: sdkmath.LegacyMustNewDecFromStr("1.1106814727683936"),
			472: sdkmath.LegacyMustNewDecFromStr("1.1109147705303473"),
			473: sdkmath.LegacyMustNewDecFromStr("1.1111483631454906"),
			474: sdkmath.LegacyMustNewDecFromStr("1.1113789833325327"),
			475: sdkmath.LegacyMustNewDecFromStr("1.1115865207841595"),
			476: sdkmath.LegacyMustNewDecFromStr("1.1118256565192843"),
			477: sdkmath.LegacyMustNewDecFromStr("1.112062977242558"),
			478: sdkmath.LegacyMustNewDecFromStr("1.112305089405149"),
			479: sdkmath.LegacyMustNewDecFromStr("1.1125496812740654"),
			480: sdkmath.LegacyMustNewDecFromStr("1.112796928321449"),
			481: sdkmath.LegacyMustNewDecFromStr("1.113045979582398"),
			482: sdkmath.LegacyMustNewDecFromStr("1.1133578645679472"),
			483: sdkmath.LegacyMustNewDecFromStr("1.1135463131500978"),
			484: sdkmath.LegacyMustNewDecFromStr("1.113862639530537"),
			485: sdkmath.LegacyMustNewDecFromStr("1.1140510045259582"),
			486: sdkmath.LegacyMustNewDecFromStr("1.114295573398525"),
			487: sdkmath.LegacyMustNewDecFromStr("1.1145990588175787"),
			488: sdkmath.LegacyMustNewDecFromStr("1.114779498371232"),
			489: sdkmath.LegacyMustNewDecFromStr("1.1150839991290917"),
			498: sdkmath.LegacyMustNewDecFromStr("1.1170896901082266"),
			499: sdkmath.LegacyMustNewDecFromStr("1.1498981693771557"),
			500: sdkmath.LegacyMustNewDecFromStr("1.209508137205966"),
			501: sdkmath.LegacyMustNewDecFromStr("1.209985009275008"),
			502: sdkmath.LegacyMustNewDecFromStr("1.210478332327813"),
			503: sdkmath.LegacyMustNewDecFromStr("1.2109676716098068"),
			504: sdkmath.LegacyMustNewDecFromStr("1.2130924701151315"),
			505: sdkmath.LegacyMustNewDecFromStr("1.2136053525521355"),
			507: sdkmath.LegacyMustNewDecFromStr("1.21455566769327"),
		},
		"evmos_9001-2": {
			499: sdkmath.LegacyMustNewDecFromStr("1.4895991845634247"),
			500: sdkmath.LegacyMustNewDecFromStr("1.4895991845634247"),
			501: sdkmath.LegacyMustNewDecFromStr("1.490098715761824"),
			502: sdkmath.LegacyMustNewDecFromStr("1.490098715761824"),
			503: sdkmath.LegacyMustNewDecFromStr("1.490098715761824"),
			504: sdkmath.LegacyMustNewDecFromStr("1.4910458236916064"),
			505: sdkmath.LegacyMustNewDecFromStr("1.4910458236916064"),
			507: sdkmath.LegacyMustNewDecFromStr("1.4918520366929944"),
			508: sdkmath.LegacyMustNewDecFromStr("1.4918520366929944"),
		},
		"osmosis-1": {
			498: sdkmath.LegacyMustNewDecFromStr("1.1984190041836773"),
			499: sdkmath.LegacyMustNewDecFromStr("1.1984190041836773"),
			500: sdkmath.LegacyMustNewDecFromStr("1.1984190041836773"),
			501: sdkmath.LegacyMustNewDecFromStr("1.1991174772238702"),
			502: sdkmath.LegacyMustNewDecFromStr("1.1991174772238702"),
			503: sdkmath.LegacyMustNewDecFromStr("1.1991174772238702"),
			504: sdkmath.LegacyMustNewDecFromStr("1.2003177583397713"),
			505: sdkmath.LegacyMustNewDecFromStr("1.2003177583397713"),
			506: sdkmath.LegacyMustNewDecFromStr("1.2003177583397713"),
			507: sdkmath.LegacyMustNewDecFromStr("1.2011986371246357"),
			508: sdkmath.LegacyMustNewDecFromStr("1.2011986371246357"),
			509: sdkmath.LegacyMustNewDecFromStr("1.2011986371246357"),
		},
		"stargaze-1": {
			498: sdkmath.LegacyMustNewDecFromStr("1.4246347073913794"),
			499: sdkmath.LegacyMustNewDecFromStr("1.4246347073913794"),
			500: sdkmath.LegacyMustNewDecFromStr("1.4246347073913794"),
			501: sdkmath.LegacyMustNewDecFromStr("1.4267297754925006"),
			502: sdkmath.LegacyMustNewDecFromStr("1.4267297754925006"),
			503: sdkmath.LegacyMustNewDecFromStr("1.4267297754925006"),
			504: sdkmath.LegacyMustNewDecFromStr("1.4279528400269015"),
			505: sdkmath.LegacyMustNewDecFromStr("1.4279528400269015"),
			506: sdkmath.LegacyMustNewDecFromStr("1.4279528400269015"),
			507: sdkmath.LegacyMustNewDecFromStr("1.430136789416802"),
			508: sdkmath.LegacyMustNewDecFromStr("1.430136789416802"),
			509: sdkmath.LegacyMustNewDecFromStr("1.430136789416802"),
		},
		"umee-1": {
			505: sdkmath.LegacyMustNewDecFromStr("1.1266406527137283"),
		},
	}
)
