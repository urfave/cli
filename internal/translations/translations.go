package translations

//go:generate gotext -srclang=en-US update -out=catalog.go -lang=en-US,en-GB github.com/urfave/cli/v3

// the languages to be supported need to be provided with the -lang flag above
// -lang=en-US,en-GB
// https://www.fincher.org/Utilities/CountryLanguageList.shtml
