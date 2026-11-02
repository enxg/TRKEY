package main

var winVK = map[string]string{
	"E00": "OEM_3",      // `
	"E11": "OEM_MINUS",  // -
	"E12": "OEM_PLUS",   // =
	"D11": "OEM_4",      // [
	"D12": "OEM_6",      // ]
	"C10": "OEM_1",      // ;
	"C11": "OEM_7",      // '
	"C12": "OEM_5",      // \ (ISO, next to Enter)
	"B00": "OEM_102",    // ISO extra key
	"B08": "OEM_COMMA",  // ,
	"B09": "OEM_PERIOD", // .
	"B10": "OEM_2",      // /
	"A03": "SPACE",      // space
}

var xkbName = map[string]string{
	"E00": "TLDE",
	"C12": "BKSL",
	"B00": "LSGT",
	"A03": "SPCE",
}

var keyName = map[string]string{
	"A03": "space",
	"E00": "grave",
	"E11": "minus",
	"E12": "equal",
	"D11": "lbracket",
	"D12": "rbracket",
	"C10": "semicolon",
	"C11": "quote",
	"C12": "backslash",
	"B00": "iso",
	"B08": "comma",
	"B09": "period",
	"B10": "slash",
}
