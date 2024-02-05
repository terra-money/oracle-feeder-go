package config

var PHOENIX_GRPC = "terra-grpc.polkachu.com:11790"
var MIGALOO_GRPC = "migaloo-grpc.polkachu.com:20790"
var KUJIRA_GRPC = "kujira-grpc.polkachu.com:11890"
var CARBON_GRPC = "carbon-grpc.polkachu.com:19690"

var AllianceDefaultConfig = AllianceConfig{
	GRPCUrls: []string{MIGALOO_GRPC, KUJIRA_GRPC, CARBON_GRPC},
	LSTSData: []LSTData{
		// Whale
		{ // Eris Protocol ampLUNA https://chainsco.pe/terra2/address/terra1ecgazyd0waaj3g7l9cmy5gulhxkps2gmxu9ghducvuypjq68mq2s5lvsct
			Symbol:   "AMPLUNA",
			IBCDenom: "ibc/05238E98A143496C8AF2B6067BABC84503909ECE9E45FBCBAC2CBA5C889FD82A",
		},
		{ // BoneLuna https://chainsco.pe/terra2/address/terra17aj4ty4sz4yhgm08na8drc0v03v2jwr3waxcqrwhajj729zhl7zqnpc0ml
			Symbol:   "BACKBONELUNA",
			IBCDenom: "ibc/40C29143BF4153B365089E40E437B7AA819672646C45BB0A5F1E10915A0B6708",
		},
		// Carbon
		{ // Eris Protocol ampLUNA https://chainsco.pe/terra2/address/terra1ecgazyd0waaj3g7l9cmy5gulhxkps2gmxu9ghducvuypjq68mq2s5lvsct
			Symbol:   "AMPLUNA",
			IBCDenom: "ibc/62A3870B9804FC3A92EAAA1F0F3F07E089DBF76CC521466CA33F5AAA8AD42290",
		},
		{ // Stride stLuna https://app.stride.zone/
			Symbol:   "STLUNA",
			IBCDenom: "ibc/FBEE20115530F474F8BBE1460DA85437C3FBBFAF4A5DEBD71CA6B9C40559A161",
		},
	},
	LSTOnPhoenix: []LSTOnPhoenix{
		{
			CounterpartyChainId: "migaloo-1",
			LSTData: LSTData{
				Symbol:   "AMPWHALE",
				IBCDenom: "ibc/B3F639855EE7478750CC8F82072307ED6E131A8EFF20345E1D136B50C4E5EC36",
			},
		},
		{
			CounterpartyChainId: "migaloo-1",
			LSTData: LSTData{
				Symbol:   "BONEWHALE",
				IBCDenom: "ibc/517E13F14A1245D4DE8CF467ADD4DA0058974CDCC880FA6AE536DBCA1D16D84E",
			},
		},
		{
			CounterpartyChainId: "carbon-1",
			LSTData: LSTData{
				Symbol:   "URSWTH",
				IBCDenom: "ibc/0E90026619DD296AD4EF9546396F292B465BAB6B5BE00ABD6162AA1CE8E68098",
			},
		},
	},
}
