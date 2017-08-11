// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package main

import "encoding/json"

var data = `
{
	"AD": {
		"Lat": 42.546245,
		"Long": 1.601554,
		"Name": "Andorra",
		"Code": "AND"
	},
	"AE": {
		"Lat": 23.424076,
		"Long": 53.847818,
		"Name": "United Arab Emirates",
		"Code": "ARE"
	},
	"AF": {
		"Lat": 33.93911,
		"Long": 67.709953,
		"Name": "Afghanistan",
		"Code": "AFG"
	},
	"AG": {
		"Lat": 17.060816,
		"Long": -61.796428,
		"Name": "Antigua and Barbuda",
		"Code": "ATG"
	},
	"AI": {
		"Lat": 18.220554,
		"Long": -63.068615,
		"Name": "Anguilla",
		"Code": "AIA"
	},
	"AL": {
		"Lat": 41.153332,
		"Long": 20.168331,
		"Name": "Albania",
		"Code": "ALB"
	},
	"AM": {
		"Lat": 40.069099,
		"Long": 45.038189,
		"Name": "Armenia",
		"Code": "ARM"
	},
	"AN": {
		"Lat": 12.226079,
		"Long": -69.060087,
		"Name": "Netherlands Antilles"
	},
	"AO": {
		"Lat": -11.202692,
		"Long": 17.873887,
		"Name": "Angola",
		"Code": "AGO"
	},
	"AQ": {
		"Lat": -75.250973,
		"Long": -0.071389,
		"Name": "Antarctica",
		"Code": "ATA"
	},
	"AR": {
		"Lat": -38.416097,
		"Long": -63.616672,
		"Name": "Argentina",
		"Code": "ARG"
	},
	"AS": {
		"Lat": -14.270972,
		"Long": -170.132217,
		"Name": "American Samoa",
		"Code": "ASM"
	},
	"AT": {
		"Lat": 47.516231,
		"Long": 14.550072,
		"Name": "Austria",
		"Code": "AUT"
	},
	"AU": {
		"Lat": -25.274398,
		"Long": 133.775136,
		"Name": "Australia",
		"Code": "AUS"
	},
	"AW": {
		"Lat": 12.52111,
		"Long": -69.968338,
		"Name": "Aruba",
		"Code": "ABW"
	},
	"AZ": {
		"Lat": 40.143105,
		"Long": 47.576927,
		"Name": "Azerbaijan",
		"Code": "AZE"
	},
	"BA": {
		"Lat": 43.915886,
		"Long": 17.679076,
		"Name": "Bosnia and Herzegovina",
		"Code": "BIH"
	},
	"BB": {
		"Lat": 13.193887,
		"Long": -59.543198,
		"Name": "Barbados",
		"Code": "BRB"
	},
	"BD": {
		"Lat": 23.684994,
		"Long": 90.356331,
		"Name": "Bangladesh",
		"Code": "BGD"
	},
	"BE": {
		"Lat": 50.503887,
		"Long": 4.469936,
		"Name": "Belgium",
		"Code": "BEL"
	},
	"BF": {
		"Lat": 12.238333,
		"Long": -1.561593,
		"Name": "Burkina Faso",
		"Code": "BFA"
	},
	"BG": {
		"Lat": 42.733883,
		"Long": 25.48583,
		"Name": "Bulgaria",
		"Code": "BGR"
	},
	"BH": {
		"Lat": 25.930414,
		"Long": 50.637772,
		"Name": "Bahrain",
		"Code": "BHR"
	},
	"BI": {
		"Lat": -3.373056,
		"Long": 29.918886,
		"Name": "Burundi",
		"Code": "BDI"
	},
	"BJ": {
		"Lat": 9.30769,
		"Long": 2.315834,
		"Name": "Benin",
		"Code": "BEN"
	},
	"BM": {
		"Lat": 32.321384,
		"Long": -64.75737,
		"Name": "Bermuda",
		"Code": "BMU"
	},
	"BN": {
		"Lat": 4.535277,
		"Long": 114.727669,
		"Name": "Brunei",
		"Code": "BRN"
	},
	"BO": {
		"Lat": -16.290154,
		"Long": -63.588653,
		"Name": "Bolivia",
		"Code": "BOL"
	},
	"BR": {
		"Lat": -14.235004,
		"Long": -51.92528,
		"Name": "Brazil",
		"Code": "BRA"
	},
	"BS": {
		"Lat": 25.03428,
		"Long": -77.39628,
		"Name": "Bahamas",
		"Code": "BHS"
	},
	"BT": {
		"Lat": 27.514162,
		"Long": 90.433601,
		"Name": "Bhutan",
		"Code": "BTN"
	},
	"BV": {
		"Lat": -54.423199,
		"Long": 3.413194,
		"Name": "Bouvet Island",
		"Code": "BVT"
	},
	"BW": {
		"Lat": -22.328474,
		"Long": 24.684866,
		"Name": "Botswana",
		"Code": "BWA"
	},
	"BY": {
		"Lat": 53.709807,
		"Long": 27.953389,
		"Name": "Belarus",
		"Code": "BLR"
	},
	"BZ": {
		"Lat": 17.189877,
		"Long": -88.49765,
		"Name": "Belize",
		"Code": "BLZ"
	},
	"CA": {
		"Lat": 56.130366,
		"Long": -106.346771,
		"Name": "Canada",
		"Code": "CAN"
	},
	"CC": {
		"Lat": -12.164165,
		"Long": 96.870956,
		"Name": "Cocos [Keeling] Islands",
		"Code": "CCK"
	},
	"CD": {
		"Lat": -4.038333,
		"Long": 21.758664,
		"Name": "Congo [DRC]",
		"Code": "COD"
	},
	"CF": {
		"Lat": 6.611111,
		"Long": 20.939444,
		"Name": "Central African Republic",
		"Code": "CAF"
	},
	"CG": {
		"Lat": -0.228021,
		"Long": 15.827659,
		"Name": "Congo [Republic]",
		"Code": "COG"
	},
	"CH": {
		"Lat": 46.818188,
		"Long": 8.227512,
		"Name": "Switzerland",
		"Code": "CHE"
	},
	"CI": {
		"Lat": 7.539989,
		"Long": -5.54708,
		"Name": "Côte d'Ivoire",
		"Code": "CIV"
	},
	"CK": {
		"Lat": -21.236736,
		"Long": -159.777671,
		"Name": "Cook Islands",
		"Code": "COK"
	},
	"CL": {
		"Lat": -35.675147,
		"Long": -71.542969,
		"Name": "Chile",
		"Code": "CHL"
	},
	"CM": {
		"Lat": 7.369722,
		"Long": 12.354722,
		"Name": "Cameroon",
		"Code": "CMR"
	},
	"CN": {
		"Lat": 35.86166,
		"Long": 104.195397,
		"Name": "China",
		"Code": "CHN"
	},
	"CO": {
		"Lat": 4.570868,
		"Long": -74.297333,
		"Name": "Colombia",
		"Code": "COL"
	},
	"CR": {
		"Lat": 9.748917,
		"Long": -83.753428,
		"Name": "Costa Rica",
		"Code": "CRI"
	},
	"CU": {
		"Lat": 21.521757,
		"Long": -77.781167,
		"Name": "Cuba",
		"Code": "CUB"
	},
	"CV": {
		"Lat": 16.002082,
		"Long": -24.013197,
		"Name": "Cape Verde",
		"Code": "CPV"
	},
	"CX": {
		"Lat": -10.447525,
		"Long": 105.690449,
		"Name": "Christmas Island",
		"Code": "CXR"
	},
	"CY": {
		"Lat": 35.126413,
		"Long": 33.429859,
		"Name": "Cyprus",
		"Code": "CYP"
	},
	"CZ": {
		"Lat": 49.817492,
		"Long": 15.472962,
		"Name": "Czech Republic",
		"Code": "CZE"
	},
	"DE": {
		"Lat": 51.165691,
		"Long": 10.451526,
		"Name": "Germany",
		"Code": "DEU"
	},
	"DJ": {
		"Lat": 11.825138,
		"Long": 42.590275,
		"Name": "Djibouti",
		"Code": "DJI"
	},
	"DK": {
		"Lat": 56.26392,
		"Long": 9.501785,
		"Name": "Denmark",
		"Code": "DNK"
	},
	"DM": {
		"Lat": 15.414999,
		"Long": -61.370976,
		"Name": "Dominica",
		"Code": "DMA"
	},
	"DO": {
		"Lat": 18.735693,
		"Long": -70.162651,
		"Name": "Dominican Republic",
		"Code": "DOM"
	},
	"DZ": {
		"Lat": 28.033886,
		"Long": 1.659626,
		"Name": "Algeria",
		"Code": "DZA"
	},
	"EC": {
		"Lat": -1.831239,
		"Long": -78.183406,
		"Name": "Ecuador",
		"Code": "ECU"
	},
	"EE": {
		"Lat": 58.595272,
		"Long": 25.013607,
		"Name": "Estonia",
		"Code": "EST"
	},
	"EG": {
		"Lat": 26.820553,
		"Long": 30.802498,
		"Name": "Egypt",
		"Code": "EGY"
	},
	"EH": {
		"Lat": 24.215527,
		"Long": -12.885834,
		"Name": "Western Sahara",
		"Code": "ESH"
	},
	"ER": {
		"Lat": 15.179384,
		"Long": 39.782334,
		"Name": "Eritrea",
		"Code": "ERI"
	},
	"ES": {
		"Lat": 40.463667,
		"Long": -3.74922,
		"Name": "Spain",
		"Code": "ESP"
	},
	"ET": {
		"Lat": 9.145,
		"Long": 40.489673,
		"Name": "Ethiopia",
		"Code": "ETH"
	},
	"FI": {
		"Lat": 61.92411,
		"Long": 25.748151,
		"Name": "Finland",
		"Code": "FIN"
	},
	"FJ": {
		"Lat": -16.578193,
		"Long": 179.414413,
		"Name": "Fiji",
		"Code": "FJI"
	},
	"FK": {
		"Lat": -51.796253,
		"Long": -59.523613,
		"Name": "Falkland Islands [Islas Malvinas]",
		"Code": "FLK"
	},
	"FM": {
		"Lat": 7.425554,
		"Long": 150.550812,
		"Name": "Micronesia",
		"Code": "FSM"
	},
	"FO": {
		"Lat": 61.892635,
		"Long": -6.911806,
		"Name": "Faroe Islands",
		"Code": "FRO"
	},
	"FR": {
		"Lat": 46.227638,
		"Long": 2.213749,
		"Name": "France",
		"Code": "FRA"
	},
	"GA": {
		"Lat": -0.803689,
		"Long": 11.609444,
		"Name": "Gabon",
		"Code": "GAB"
	},
	"GB": {
		"Lat": 55.378051,
		"Long": -3.435973,
		"Name": "United Kingdom",
		"Code": "GBR"
	},
	"GD": {
		"Lat": 12.262776,
		"Long": -61.604171,
		"Name": "Grenada",
		"Code": "GRD"
	},
	"GE": {
		"Lat": 42.315407,
		"Long": 43.356892,
		"Name": "Georgia",
		"Code": "GEO"
	},
	"GF": {
		"Lat": 3.933889,
		"Long": -53.125782,
		"Name": "French Guiana",
		"Code": "GUF"
	},
	"GG": {
		"Lat": 49.465691,
		"Long": -2.585278,
		"Name": "Guernsey",
		"Code": "GGY"
	},
	"GH": {
		"Lat": 7.946527,
		"Long": -1.023194,
		"Name": "Ghana",
		"Code": "GHA"
	},
	"GI": {
		"Lat": 36.137741,
		"Long": -5.345374,
		"Name": "Gibraltar",
		"Code": "GIB"
	},
	"GL": {
		"Lat": 71.706936,
		"Long": -42.604303,
		"Name": "Greenland",
		"Code": "GRL"
	},
	"GM": {
		"Lat": 13.443182,
		"Long": -15.310139,
		"Name": "Gambia",
		"Code": "GMB"
	},
	"GN": {
		"Lat": 9.945587,
		"Long": -9.696645,
		"Name": "Guinea",
		"Code": "GIN"
	},
	"GP": {
		"Lat": 16.995971,
		"Long": -62.067641,
		"Name": "Guadeloupe",
		"Code": "GLP"
	},
	"GQ": {
		"Lat": 1.650801,
		"Long": 10.267895,
		"Name": "Equatorial Guinea",
		"Code": "GNQ"
	},
	"GR": {
		"Lat": 39.074208,
		"Long": 21.824312,
		"Name": "Greece",
		"Code": "GRC"
	},
	"GS": {
		"Lat": -54.429579,
		"Long": -36.587909,
		"Name": "South Georgia and the South Sandwich Islands",
		"Code": "SGS"
	},
	"GT": {
		"Lat": 15.783471,
		"Long": -90.230759,
		"Name": "Guatemala",
		"Code": "GTM"
	},
	"GU": {
		"Lat": 13.444304,
		"Long": 144.793731,
		"Name": "Guam",
		"Code": "GUM"
	},
	"GW": {
		"Lat": 11.803749,
		"Long": -15.180413,
		"Name": "Guinea-Bissau",
		"Code": "GNB"
	},
	"GY": {
		"Lat": 4.860416,
		"Long": -58.93018,
		"Name": "Guyana",
		"Code": "GUY"
	},
	"GZ": {
		"Lat": 31.354676,
		"Long": 34.308825,
		"Name": "Gaza Strip"
	},
	"HK": {
		"Lat": 22.396428,
		"Long": 114.109497,
		"Name": "Hong Kong",
		"Code": "HKG"
	},
	"HM": {
		"Lat": -53.08181,
		"Long": 73.504158,
		"Name": "Heard Island and McDonald Islands",
		"Code": "HMD"
	},
	"HN": {
		"Lat": 15.199999,
		"Long": -86.241905,
		"Name": "Honduras",
		"Code": "HND"
	},
	"HR": {
		"Lat": 45.1,
		"Long": 15.2,
		"Name": "Croatia",
		"Code": "HRV"
	},
	"HT": {
		"Lat": 18.971187,
		"Long": -72.285215,
		"Name": "Haiti",
		"Code": "HTI"
	},
	"HU": {
		"Lat": 47.162494,
		"Long": 19.503304,
		"Name": "Hungary",
		"Code": "HUN"
	},
	"ID": {
		"Lat": -0.789275,
		"Long": 113.921327,
		"Name": "Indonesia",
		"Code": "IDN"
	},
	"IE": {
		"Lat": 53.41291,
		"Long": -8.24389,
		"Name": "Ireland",
		"Code": "IRL"
	},
	"IL": {
		"Lat": 31.046051,
		"Long": 34.851612,
		"Name": "Israel",
		"Code": "ISR"
	},
	"IM": {
		"Lat": 54.236107,
		"Long": -4.548056,
		"Name": "Isle of Man",
		"Code": "IMN"
	},
	"IN": {
		"Lat": 20.593684,
		"Long": 78.96288,
		"Name": "India",
		"Code": "IND"
	},
	"IO": {
		"Lat": -6.343194,
		"Long": 71.876519,
		"Name": "British Indian Ocean Territory",
		"Code": "IOT"
	},
	"IQ": {
		"Lat": 33.223191,
		"Long": 43.679291,
		"Name": "Iraq",
		"Code": "IRQ"
	},
	"IR": {
		"Lat": 32.427908,
		"Long": 53.688046,
		"Name": "Iran",
		"Code": "IRN"
	},
	"IS": {
		"Lat": 64.963051,
		"Long": -19.020835,
		"Name": "Iceland",
		"Code": "ISL"
	},
	"IT": {
		"Lat": 41.87194,
		"Long": 12.56738,
		"Name": "Italy",
		"Code": "ITA"
	},
	"JE": {
		"Lat": 49.214439,
		"Long": -2.13125,
		"Name": "Jersey",
		"Code": "JEY"
	},
	"JM": {
		"Lat": 18.109581,
		"Long": -77.297508,
		"Name": "Jamaica",
		"Code": "JAM"
	},
	"JO": {
		"Lat": 30.585164,
		"Long": 36.238414,
		"Name": "Jordan",
		"Code": "JOR"
	},
	"JP": {
		"Lat": 36.204824,
		"Long": 138.252924,
		"Name": "Japan",
		"Code": "JPN"
	},
	"KE": {
		"Lat": -0.023559,
		"Long": 37.906193,
		"Name": "Kenya",
		"Code": "KEN"
	},
	"KG": {
		"Lat": 41.20438,
		"Long": 74.766098,
		"Name": "Kyrgyzstan",
		"Code": "KGZ"
	},
	"KH": {
		"Lat": 12.565679,
		"Long": 104.990963,
		"Name": "Cambodia",
		"Code": "KHM"
	},
	"KI": {
		"Lat": -3.370417,
		"Long": -168.734039,
		"Name": "Kiribati",
		"Code": "KIR"
	},
	"KM": {
		"Lat": -11.875001,
		"Long": 43.872219,
		"Name": "Comoros",
		"Code": "COM"
	},
	"KN": {
		"Lat": 17.357822,
		"Long": -62.782998,
		"Name": "Saint Kitts and Nevis",
		"Code": "KNA"
	},
	"KP": {
		"Lat": 40.339852,
		"Long": 127.510093,
		"Name": "North Korea",
		"Code": "PRK"
	},
	"KR": {
		"Lat": 35.907757,
		"Long": 127.766922,
		"Name": "South Korea",
		"Code": "KOR"
	},
	"KW": {
		"Lat": 29.31166,
		"Long": 47.481766,
		"Name": "Kuwait",
		"Code": "KWT"
	},
	"KY": {
		"Lat": 19.513469,
		"Long": -80.566956,
		"Name": "Cayman Islands",
		"Code": "CYM"
	},
	"KZ": {
		"Lat": 48.019573,
		"Long": 66.923684,
		"Name": "Kazakhstan",
		"Code": "KAZ"
	},
	"LA": {
		"Lat": 19.85627,
		"Long": 102.495496,
		"Name": "Laos",
		"Code": "LAO"
	},
	"LB": {
		"Lat": 33.854721,
		"Long": 35.862285,
		"Name": "Lebanon",
		"Code": "LBN"
	},
	"LC": {
		"Lat": 13.909444,
		"Long": -60.978893,
		"Name": "Saint Lucia",
		"Code": "LCA"
	},
	"LI": {
		"Lat": 47.166,
		"Long": 9.555373,
		"Name": "Liechtenstein",
		"Code": "LIE"
	},
	"LK": {
		"Lat": 7.873054,
		"Long": 80.771797,
		"Name": "Sri Lanka",
		"Code": "LKA"
	},
	"LR": {
		"Lat": 6.428055,
		"Long": -9.429499,
		"Name": "Liberia",
		"Code": "LBR"
	},
	"LS": {
		"Lat": -29.609988,
		"Long": 28.233608,
		"Name": "Lesotho",
		"Code": "LSO"
	},
	"LT": {
		"Lat": 55.169438,
		"Long": 23.881275,
		"Name": "Lithuania",
		"Code": "LTU"
	},
	"LU": {
		"Lat": 49.815273,
		"Long": 6.129583,
		"Name": "Luxembourg",
		"Code": "LUX"
	},
	"LV": {
		"Lat": 56.879635,
		"Long": 24.603189,
		"Name": "Latvia",
		"Code": "LVA"
	},
	"LY": {
		"Lat": 26.3351,
		"Long": 17.228331,
		"Name": "Libya",
		"Code": "LBY"
	},
	"MA": {
		"Lat": 31.791702,
		"Long": -7.09262,
		"Name": "Morocco",
		"Code": "MAR"
	},
	"MC": {
		"Lat": 43.750298,
		"Long": 7.412841,
		"Name": "Monaco",
		"Code": "MCO"
	},
	"MD": {
		"Lat": 47.411631,
		"Long": 28.369885,
		"Name": "Moldova",
		"Code": "MDA"
	},
	"ME": {
		"Lat": 42.708678,
		"Long": 19.37439,
		"Name": "Montenegro",
		"Code": "MNE"
	},
	"MG": {
		"Lat": -18.766947,
		"Long": 46.869107,
		"Name": "Madagascar",
		"Code": "MDG"
	},
	"MH": {
		"Lat": 7.131474,
		"Long": 171.184478,
		"Name": "Marshall Islands",
		"Code": "MHL"
	},
	"MK": {
		"Lat": 41.608635,
		"Long": 21.745275,
		"Name": "Macedonia [FYROM]",
		"Code": "MKD"
	},
	"ML": {
		"Lat": 17.570692,
		"Long": -3.996166,
		"Name": "Mali",
		"Code": "MLI"
	},
	"MM": {
		"Lat": 21.913965,
		"Long": 95.956223,
		"Name": "Myanmar [Burma]",
		"Code": "MMR"
	},
	"MN": {
		"Lat": 46.862496,
		"Long": 103.846656,
		"Name": "Mongolia",
		"Code": "MNG"
	},
	"MO": {
		"Lat": 22.198745,
		"Long": 113.543873,
		"Name": "Macau",
		"Code": "MAC"
	},
	"MP": {
		"Lat": 17.33083,
		"Long": 145.38469,
		"Name": "Northern Mariana Islands",
		"Code": "MNP"
	},
	"MQ": {
		"Lat": 14.641528,
		"Long": -61.024174,
		"Name": "Martinique",
		"Code": "MTQ"
	},
	"MR": {
		"Lat": 21.00789,
		"Long": -10.940835,
		"Name": "Mauritania",
		"Code": "MRT"
	},
	"MS": {
		"Lat": 16.742498,
		"Long": -62.187366,
		"Name": "Montserrat",
		"Code": "MSR"
	},
	"MT": {
		"Lat": 35.937496,
		"Long": 14.375416,
		"Name": "Malta",
		"Code": "MLT"
	},
	"MU": {
		"Lat": -20.348404,
		"Long": 57.552152,
		"Name": "Mauritius",
		"Code": "MUS"
	},
	"MV": {
		"Lat": 3.202778,
		"Long": 73.22068,
		"Name": "Maldives",
		"Code": "MDV"
	},
	"MW": {
		"Lat": -13.254308,
		"Long": 34.301525,
		"Name": "Malawi",
		"Code": "MWI"
	},
	"MX": {
		"Lat": 23.634501,
		"Long": -102.552784,
		"Name": "Mexico",
		"Code": "MEX"
	},
	"MY": {
		"Lat": 4.210484,
		"Long": 101.975766,
		"Name": "Malaysia",
		"Code": "MYS"
	},
	"MZ": {
		"Lat": -18.665695,
		"Long": 35.529562,
		"Name": "Mozambique",
		"Code": "MOZ"
	},
	"NA": {
		"Lat": -22.95764,
		"Long": 18.49041,
		"Name": "Namibia",
		"Code": "NAM"
	},
	"NC": {
		"Lat": -20.904305,
		"Long": 165.618042,
		"Name": "New Caledonia",
		"Code": "NCL"
	},
	"NE": {
		"Lat": 17.607789,
		"Long": 8.081666,
		"Name": "Niger",
		"Code": "NER"
	},
	"NF": {
		"Lat": -29.040835,
		"Long": 167.954712,
		"Name": "Norfolk Island",
		"Code": "NFK"
	},
	"NG": {
		"Lat": 9.081999,
		"Long": 8.675277,
		"Name": "Nigeria",
		"Code": "NGA"
	},
	"NI": {
		"Lat": 12.865416,
		"Long": -85.207229,
		"Name": "Nicaragua",
		"Code": "NIC"
	},
	"NL": {
		"Lat": 52.132633,
		"Long": 5.291266,
		"Name": "Netherlands",
		"Code": "NLD"
	},
	"NO": {
		"Lat": 60.472024,
		"Long": 8.468946,
		"Name": "Norway",
		"Code": "NOR"
	},
	"NP": {
		"Lat": 28.394857,
		"Long": 84.124008,
		"Name": "Nepal",
		"Code": "NPL"
	},
	"NR": {
		"Lat": -0.522778,
		"Long": 166.931503,
		"Name": "Nauru",
		"Code": "NRU"
	},
	"NU": {
		"Lat": -19.054445,
		"Long": -169.867233,
		"Name": "Niue",
		"Code": "NIU"
	},
	"NZ": {
		"Lat": -40.900557,
		"Long": 174.885971,
		"Name": "New Zealand",
		"Code": "NZL"
	},
	"OM": {
		"Lat": 21.512583,
		"Long": 55.923255,
		"Name": "Oman",
		"Code": "OMN"
	},
	"PA": {
		"Lat": 8.537981,
		"Long": -80.782127,
		"Name": "Panama",
		"Code": "PAN"
	},
	"PE": {
		"Lat": -9.189967,
		"Long": -75.015152,
		"Name": "Peru",
		"Code": "PER"
	},
	"PF": {
		"Lat": -17.679742,
		"Long": -149.406843,
		"Name": "French Polynesia",
		"Code": "PYF"
	},
	"PG": {
		"Lat": -6.314993,
		"Long": 143.95555,
		"Name": "Papua New Guinea",
		"Code": "PNG"
	},
	"PH": {
		"Lat": 12.879721,
		"Long": 121.774017,
		"Name": "Philippines",
		"Code": "PHL"
	},
	"PK": {
		"Lat": 30.375321,
		"Long": 69.345116,
		"Name": "Pakistan",
		"Code": "PAK"
	},
	"PL": {
		"Lat": 51.919438,
		"Long": 19.145136,
		"Name": "Poland",
		"Code": "POL"
	},
	"PM": {
		"Lat": 46.941936,
		"Long": -56.27111,
		"Name": "Saint Pierre and Miquelon",
		"Code": "SPM"
	},
	"PN": {
		"Lat": -24.703615,
		"Long": -127.439308,
		"Name": "Pitcairn Islands",
		"Code": "PCN"
	},
	"PR": {
		"Lat": 18.220833,
		"Long": -66.590149,
		"Name": "Puerto Rico",
		"Code": "PRI"
	},
	"PS": {
		"Lat": 31.952162,
		"Long": 35.233154,
		"Name": "Palestinian Territories",
		"Code": "PSE"
	},
	"PT": {
		"Lat": 39.399872,
		"Long": -8.224454,
		"Name": "Portugal",
		"Code": "PRT"
	},
	"PW": {
		"Lat": 7.51498,
		"Long": 134.58252,
		"Name": "Palau",
		"Code": "PLW"
	},
	"PY": {
		"Lat": -23.442503,
		"Long": -58.443832,
		"Name": "Paraguay",
		"Code": "PRY"
	},
	"QA": {
		"Lat": 25.354826,
		"Long": 51.183884,
		"Name": "Qatar",
		"Code": "QAT"
	},
	"RE": {
		"Lat": -21.115141,
		"Long": 55.536384,
		"Name": "Réunion",
		"Code": "REU"
	},
	"RO": {
		"Lat": 45.943161,
		"Long": 24.96676,
		"Name": "Romania",
		"Code": "ROU"
	},
	"RS": {
		"Lat": 44.016521,
		"Long": 21.005859,
		"Name": "Serbia",
		"Code": "SRB"
	},
	"RU": {
		"Lat": 61.52401,
		"Long": 105.318756,
		"Name": "Russia",
		"Code": "RUS"
	},
	"RW": {
		"Lat": -1.940278,
		"Long": 29.873888,
		"Name": "Rwanda",
		"Code": "RWA"
	},
	"SA": {
		"Lat": 23.885942,
		"Long": 45.079162,
		"Name": "Saudi Arabia",
		"Code": "SAU"
	},
	"SB": {
		"Lat": -9.64571,
		"Long": 160.156194,
		"Name": "Solomon Islands",
		"Code": "SLB"
	},
	"SC": {
		"Lat": -4.679574,
		"Long": 55.491977,
		"Name": "Seychelles",
		"Code": "SYC"
	},
	"SD": {
		"Lat": 12.862807,
		"Long": 30.217636,
		"Name": "Sudan",
		"Code": "SDN"
	},
	"SE": {
		"Lat": 60.128161,
		"Long": 18.643501,
		"Name": "Sweden",
		"Code": "SWE"
	},
	"SG": {
		"Lat": 1.352083,
		"Long": 103.819836,
		"Name": "Singapore",
		"Code": "SGP"
	},
	"SH": {
		"Lat": -24.143474,
		"Long": -10.030696,
		"Name": "Saint Helena",
		"Code": "SHN"
	},
	"SI": {
		"Lat": 46.151241,
		"Long": 14.995463,
		"Name": "Slovenia",
		"Code": "SVN"
	},
	"SJ": {
		"Lat": 77.553604,
		"Long": 23.670272,
		"Name": "Svalbard and Jan Mayen",
		"Code": "SJM"
	},
	"SK": {
		"Lat": 48.669026,
		"Long": 19.699024,
		"Name": "Slovakia",
		"Code": "SVK"
	},
	"SL": {
		"Lat": 8.460555,
		"Long": -11.779889,
		"Name": "Sierra Leone",
		"Code": "SLE"
	},
	"SM": {
		"Lat": 43.94236,
		"Long": 12.457777,
		"Name": "San Marino",
		"Code": "SMR"
	},
	"SN": {
		"Lat": 14.497401,
		"Long": -14.452362,
		"Name": "Senegal",
		"Code": "SEN"
	},
	"SO": {
		"Lat": 5.152149,
		"Long": 46.199616,
		"Name": "Somalia",
		"Code": "SOM"
	},
	"SR": {
		"Lat": 3.919305,
		"Long": -56.027783,
		"Name": "Suriname",
		"Code": "SUR"
	},
	"ST": {
		"Lat": 0.18636,
		"Long": 6.613081,
		"Name": "São Tomé and Príncipe",
		"Code": "STP"
	},
	"SV": {
		"Lat": 13.794185,
		"Long": -88.89653,
		"Name": "El Salvador",
		"Code": "SLV"
	},
	"SY": {
		"Lat": 34.802075,
		"Long": 38.996815,
		"Name": "Syria",
		"Code": "SYR"
	},
	"SZ": {
		"Lat": -26.522503,
		"Long": 31.465866,
		"Name": "Swaziland",
		"Code": "SWZ"
	},
	"TC": {
		"Lat": 21.694025,
		"Long": -71.797928,
		"Name": "Turks and Caicos Islands",
		"Code": "TCA"
	},
	"TD": {
		"Lat": 15.454166,
		"Long": 18.732207,
		"Name": "Chad",
		"Code": "TCD"
	},
	"TF": {
		"Lat": -49.280366,
		"Long": 69.348557,
		"Name": "French Southern Territories",
		"Code": "ATF"
	},
	"TG": {
		"Lat": 8.619543,
		"Long": 0.824782,
		"Name": "Togo",
		"Code": "TGO"
	},
	"TH": {
		"Lat": 15.870032,
		"Long": 100.992541,
		"Name": "Thailand",
		"Code": "THA"
	},
	"TJ": {
		"Lat": 38.861034,
		"Long": 71.276093,
		"Name": "Tajikistan",
		"Code": "TJK"
	},
	"TK": {
		"Lat": -8.967363,
		"Long": -171.855881,
		"Name": "Tokelau",
		"Code": "TKL"
	},
	"TL": {
		"Lat": -8.874217,
		"Long": 125.727539,
		"Name": "Timor-Leste",
		"Code": "TLS"
	},
	"TM": {
		"Lat": 38.969719,
		"Long": 59.556278,
		"Name": "Turkmenistan",
		"Code": "TKM"
	},
	"TN": {
		"Lat": 33.886917,
		"Long": 9.537499,
		"Name": "Tunisia",
		"Code": "TUN"
	},
	"TO": {
		"Lat": -21.178986,
		"Long": -175.198242,
		"Name": "Tonga",
		"Code": "TON"
	},
	"TR": {
		"Lat": 38.963745,
		"Long": 35.243322,
		"Name": "Turkey",
		"Code": "TUR"
	},
	"TT": {
		"Lat": 10.691803,
		"Long": -61.222503,
		"Name": "Trinidad and Tobago",
		"Code": "TTO"
	},
	"TV": {
		"Lat": -7.109535,
		"Long": 177.64933,
		"Name": "Tuvalu",
		"Code": "TUV"
	},
	"TW": {
		"Lat": 23.69781,
		"Long": 120.960515,
		"Name": "Taiwan",
		"Code": "TWN"
	},
	"TZ": {
		"Lat": -6.369028,
		"Long": 34.888822,
		"Name": "Tanzania",
		"Code": "TZA"
	},
	"UA": {
		"Lat": 48.379433,
		"Long": 31.16558,
		"Name": "Ukraine",
		"Code": "UKR"
	},
	"UG": {
		"Lat": 1.373333,
		"Long": 32.290275,
		"Name": "Uganda",
		"Code": "UGA"
	},
	"UM": {
		"Lat": "",
		"Long": "",
		"Name": "U.S. Minor Outlying Islands",
		"Code": "UMI"
	},
	"US": {
		"Lat": 37.09024,
		"Long": -95.712891,
		"Name": "United States",
		"Code": "USA"
	},
	"UY": {
		"Lat": -32.522779,
		"Long": -55.765835,
		"Name": "Uruguay",
		"Code": "URY"
	},
	"UZ": {
		"Lat": 41.377491,
		"Long": 64.585262,
		"Name": "Uzbekistan",
		"Code": "UZB"
	},
	"VA": {
		"Lat": 41.902916,
		"Long": 12.453389,
		"Name": "Vatican City",
		"Code": "VAT"
	},
	"VC": {
		"Lat": 12.984305,
		"Long": -61.287228,
		"Name": "Saint Vincent and the Grenadines",
		"Code": "VCT"
	},
	"VE": {
		"Lat": 6.42375,
		"Long": -66.58973,
		"Name": "Venezuela",
		"Code": "VEN"
	},
	"VG": {
		"Lat": 18.420695,
		"Long": -64.639968,
		"Name": "British Virgin Islands",
		"Code": "VGB"
	},
	"VI": {
		"Lat": 18.335765,
		"Long": -64.896335,
		"Name": "U.S. Virgin Islands",
		"Code": "VIR"
	},
	"VN": {
		"Lat": 14.058324,
		"Long": 108.277199,
		"Name": "Vietnam",
		"Code": "VNM"
	},
	"VU": {
		"Lat": -15.376706,
		"Long": 166.959158,
		"Name": "Vanuatu",
		"Code": "VUT"
	},
	"WF": {
		"Lat": -13.768752,
		"Long": -177.156097,
		"Name": "Wallis and Futuna",
		"Code": "WLF"
	},
	"WS": {
		"Lat": -13.759029,
		"Long": -172.104629,
		"Name": "Samoa",
		"Code": "WSM"
	},
	"XK": {
		"Lat": 42.602636,
		"Long": 20.902977,
		"Name": "Kosovo",
		"Code": "XKX"
	},
	"YE": {
		"Lat": 15.552727,
		"Long": 48.516388,
		"Name": "Yemen",
		"Code": "YEM"
	},
	"YT": {
		"Lat": -12.8275,
		"Long": 45.166244,
		"Name": "Mayotte",
		"Code": "MYT"
	},
	"ZA": {
		"Lat": -30.559482,
		"Long": 22.937506,
		"Name": "South Africa",
		"Code": "ZAF"
	},
	"ZM": {
		"Lat": -13.133897,
		"Long": 27.849332,
		"Name": "Zambia",
		"Code": "ZMB"
	},
	"ZW": {
		"Lat": -19.015438,
		"Long": 29.154857,
		"Name": "Zimbabwe",
		"Code": "ZWE"
	}
}`

type country struct {
	Name string
	Code string
	Lat  float32
	Long float32
}

func countries() map[string]country {
	m := make(map[string]country, 0)
	json.Unmarshal([]byte(data), &m)
	return m
}
