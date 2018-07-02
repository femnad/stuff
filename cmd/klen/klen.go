package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/femnad/mare"
)

// Multiples of bytes in units of powers of ten
const (
	PiB = 5.0
	TiB = 4.0
	GiB = 3.0
	MiB = 2.0
	KiB = 1.0
	B   = 0.0

	defaultPrecision = 2
	twoToTen         = 1024
	usage            = "%s: simpleton binary multiple converter\nusage:\n"
)

var units = map[string]float64{
	"P": PiB,
	"T": TiB,
	"G": GiB,
	"M": MiB,
	"K": KiB,
	"B": B,
}

func convert(input, output, number string) float64 {
	parsedNumber, err := strconv.ParseFloat(number, 64)
	mare.PanicIfErr(err)

	inUnit := math.Pow(twoToTen, units[input])
	outUnit := math.Pow(twoToTen, units[output])

	return inUnit * parsedNumber / outUnit
}

func getFormatString(numberOfDigits int) string {
	return fmt.Sprintf("%%.%df\n", numberOfDigits)
}

func getFlagSet(name string, arguments []string) *flag.FlagSet {
	flagSet := flag.NewFlagSet(name, flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintf(flagSet.Output(), usage, name)
		flagSet.PrintDefaults()
	}
	return flagSet
}

func parseFlags(flagSet *flag.FlagSet, inputUnit, outputUnit string, precisionPoints int) string {
	flagSet.StringVar(&inputUnit, "i", "B", "input unit")
	flagSet.StringVar(&outputUnit, "o", "G", "output unit")

	flagSet.IntVar(&precisionPoints, "p", defaultPrecision, "number of decimal precision digits")

	flagSet.Parse(os.Args[1:])
	number := flagSet.Arg(0)

	if number == "" {
		flagSet.PrintDefaults()
		os.Exit(1)
	}

	return number
}

func main() {
	flagSet := getFlagSet(os.Args[0], os.Args[1:])

	var inputUnit, outputUnit string
	var precisionPoints int

	number := parseFlags(flagSet, inputUnit, outputUnit, precisionPoints)

	converted := convert(inputUnit, outputUnit, number)
	formatString := getFormatString(precisionPoints)

	fmt.Printf(formatString, converted)
}
