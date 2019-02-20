package mmdatformat

import (
	"errors"
	"strings"

	"github.com/anexia-it/geodbtools"
)

var countryCodesISO2Mappings = map[string]string{
	"XK": "RS",
}

var countryCodesISO2 = []string{
	"", "AP", "EU", "AD", "AE", "AF",
	"AG", "AI", "AL", "AM", "CW",
	"AO", "AQ", "AR", "AS", "AT", "AU",
	"AW", "AZ", "BA", "BB",
	"BD", "BE", "BF", "BG", "BH", "BI",
	"BJ", "BM", "BN", "BO",
	"BR", "BS", "BT", "BV", "BW", "BY",
	"BZ", "CA", "CC", "CD",
	"CF", "CG", "CH", "CI", "CK", "CL",
	"CM", "CN", "CO", "CR",
	"CU", "CV", "CX", "CY", "CZ", "DE",
	"DJ", "DK", "DM", "DO",
	"DZ", "EC", "EE", "EG", "EH", "ER",
	"ES", "ET", "FI", "FJ",
	"FK", "FM", "FO", "FR", "SX", "GA",
	"GB", "GD", "GE", "GF",
	"GH", "GI", "GL", "GM", "GN", "GP",
	"GQ", "GR", "GS", "GT",
	"GU", "GW", "GY", "HK", "HM", "HN",
	"HR", "HT", "HU", "ID",
	"IE", "IL", "IN", "IO", "IQ", "IR",
	"IS", "IT", "JM", "JO",
	"JP", "KE", "KG", "KH", "KI", "KM",
	"KN", "KP", "KR", "KW",
	"KY", "KZ", "LA", "LB", "LC", "LI",
	"LK", "LR", "LS", "LT",
	"LU", "LV", "LY", "MA", "MC", "MD",
	"MG", "MH", "MK", "ML",
	"MM", "MN", "MO", "MP", "MQ", "MR",
	"MS", "MT", "MU", "MV",
	"MW", "MX", "MY", "MZ", "NA", "NC",
	"NE", "NF", "NG", "NI",
	"NL", "NO", "NP", "NR", "NU", "NZ",
	"OM", "PA", "PE", "PF",
	"PG", "PH", "PK", "PL", "PM", "PN",
	"PR", "PS", "PT", "PW",
	"PY", "QA", "RE", "RO", "RU", "RW",
	"SA", "SB", "SC", "SD",
	"SE", "SG", "SH", "SI", "SJ", "SK",
	"SL", "SM", "SN", "SO",
	"SR", "ST", "SV", "SY", "SZ", "TC",
	"TD", "TF", "TG", "TH",
	"TJ", "TK", "TM", "TN", "TO", "TL",
	"TR", "TT", "TV", "TW",
	"TZ", "UA", "UG", "UM", "US", "UY",
	"UZ", "VA", "VC", "VE",
	"VG", "VI", "VN", "VU", "WF", "WS",
	"YE", "YT", "RS", "ZA",
	"ZM", "ME", "ZW", "A1", "A2", "O1",
	"AX", "GG", "IM", "JE",
	"BL", "MF", "BQ", "SS", "O1",
}

var countryCodesISO3Mappings = map[string]string{
	"RKS": "SRB",
}

var countryCodesISO3 = []string{
	"--", "AP", "EU", "AND",
	"ARE",
	"AFG", "ATG", "AIA", "ALB", "ARM", "CUW",
	"AGO", "ATA", "ARG", "ASM", "AUT",
	"AUS", "ABW", "AZE", "BIH", "BRB",
	"BGD", "BEL", "BFA", "BGR", "BHR",
	"BDI", "BEN", "BMU", "BRN", "BOL",
	"BRA", "BHS", "BTN", "BVT", "BWA",
	"BLR", "BLZ", "CAN", "CCK", "COD",
	"CAF", "COG", "CHE", "CIV", "COK",
	"CHL", "CMR", "CHN", "COL", "CRI",
	"CUB", "CPV", "CXR", "CYP", "CZE",
	"DEU", "DJI", "DNK", "DMA", "DOM",
	"DZA", "ECU", "EST", "EGY", "ESH",
	"ERI", "ESP", "ETH", "FIN", "FJI",
	"FLK", "FSM", "FRO", "FRA", "SXM",
	"GAB", "GBR", "GRD", "GEO", "GUF",
	"GHA", "GIB", "GRL", "GMB", "GIN",
	"GLP", "GNQ", "GRC", "SGS", "GTM",
	"GUM", "GNB", "GUY", "HKG", "HMD",
	"HND", "HRV", "HTI", "HUN", "IDN",
	"IRL", "ISR", "IND", "IOT", "IRQ",
	"IRN", "ISL", "ITA", "JAM", "JOR",
	"JPN", "KEN", "KGZ", "KHM", "KIR",
	"COM", "KNA", "PRK", "KOR", "KWT",
	"CYM", "KAZ", "LAO", "LBN", "LCA",
	"LIE", "LKA", "LBR", "LSO", "LTU",
	"LUX", "LVA", "LBY", "MAR", "MCO",
	"MDA", "MDG", "MHL", "MKD", "MLI",
	"MMR", "MNG", "MAC", "MNP", "MTQ",
	"MRT", "MSR", "MLT", "MUS", "MDV",
	"MWI", "MEX", "MYS", "MOZ", "NAM",
	"NCL", "NER", "NFK", "NGA", "NIC",
	"NLD", "NOR", "NPL", "NRU", "NIU",
	"NZL", "OMN", "PAN", "PER", "PYF",
	"PNG", "PHL", "PAK", "POL", "SPM",
	"PCN", "PRI", "PSE", "PRT", "PLW",
	"PRY", "QAT", "REU", "ROU", "RUS",
	"RWA", "SAU", "SLB", "SYC", "SDN",
	"SWE", "SGP", "SHN", "SVN", "SJM",
	"SVK", "SLE", "SMR", "SEN", "SOM",
	"SUR", "STP", "SLV", "SYR", "SWZ",
	"TCA", "TCD", "ATF", "TGO", "THA",
	"TJK", "TKL", "TKM", "TUN", "TON",
	"TLS", "TUR", "TTO", "TUV", "TWN",
	"TZA", "UKR", "UGA", "UMI", "USA",
	"URY", "UZB", "VAT", "VCT", "VEN",
	"VGB", "VIR", "VNM", "VUT", "WLF",
	"WSM", "YEM", "MYT", "SRB", "ZAF",
	"ZMB", "MNE", "ZWE", "A1", "A2",
	"O1", "ALA", "GGY", "IMN", "JEY",
	"BLM", "MAF", "BES", "SSD", "O1",
}

var (
	// ErrCountryNotFound indicates that a country was not found
	ErrCountryNotFound = errors.New("country not found")
)

func getCountryCodeIndex(countryCodes []string, mappings map[string]string, countryCode string) (idx int, err error) {
	idx = -1
	countryCode = strings.ToUpper(countryCode)

	if mappedCode, mappingExists := mappings[countryCode]; mappingExists {
		countryCode = mappedCode
	}

	for i, code := range countryCodes {
		if code == countryCode {
			idx = i
			break
		}
	}

	if idx < 0 {
		err = ErrCountryNotFound
	}

	return
}

// GetISO2CountryCodeIndex retrieves the index of a given GeoIP country code in 2 char format
func GetISO2CountryCodeIndex(countryCode string) (idx int, err error) {
	if countryCode == "" {
		return
	}

	if len(countryCode) != 2 {
		idx = -1
		err = ErrCountryNotFound
		return
	}

	return getCountryCodeIndex(countryCodesISO2, countryCodesISO2Mappings, countryCode)
}

// GetISO3CountryCodeIndex retrieves the index of a given GeoIP country code in 3 char format
func GetISO3CountryCodeIndex(countryCode string) (idx int, err error) {
	if countryCode == "" {
		return
	}

	if len(countryCode) < 2 || len(countryCode) > 3 {
		idx = -1
		err = ErrCountryNotFound
		return
	}

	return getCountryCodeIndex(countryCodesISO3, countryCodesISO3Mappings, countryCode)
}

func getCountryCodeString(countryCodes []string, idx int) (countryCode string, err error) {
	if idx < 0 || idx >= len(countryCodes) {
		err = ErrCountryNotFound
		return
	}

	countryCode = countryCodes[idx]
	return
}

// GetISO2CountryCodeString retrieves the 2-char country code for a given index
func GetISO2CountryCodeString(idx int) (countryCode string, err error) {
	return getCountryCodeString(countryCodesISO2, idx)
}

// GetISO3CountryCodeString retrieves the 3-char country code for a given index
func GetISO3CountryCodeString(idx int) (countryCode string, err error) {
	return getCountryCodeString(countryCodesISO3, idx)
}

func init() {
	for countryCode, equalCountryCode := range countryCodesISO2Mappings {
		geodbtools.RegisterEquivalentCountryCode(countryCode, equalCountryCode)
	}
}
