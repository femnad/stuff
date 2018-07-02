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
)

var units = map[string]float64{
	"P": PiB,
	"T": TiB,
	"G": GiB,
	"M": MiB,
	"K": KiB,
	"B": B,
}

type params struct {
	inputUnit string
	out       string
	precision int
}

func getMultiple(unit string) float64 {
	unitQuantity, ok := units[unit]
	if !ok {
		msg := fmt.Sprintf("Unable to find multiplier for unit %s", unit)
		panic(msg)
	}
	return unitQuantity
}

func powerOf1024(power float64) float64 {
	return math.Pow(twoToTen, power)
}

func convert(input, output string, number float64) float64 {
	inUnit := getMultiple(input)
	outUnit := getMultiple(output)
	return powerOf1024(inUnit) * number / powerOf1024(outUnit)
}

func getFormatString(numberOfDigits int) string {
	return fmt.Sprintf("%%.%df\n", numberOfDigits)
}

func ensureCorrectNumArgs() {
	arguments := flag.Args()
	numArgs := len(arguments)
	if numArgs != 1 {
		msg := fmt.Sprintf("Incorrect number of arguments: %d", numArgs)
		panic(msg)
	}
}

func parseNumber() float64 {
	number := flag.Arg(0)
	if number == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	parsedNumber, err := strconv.ParseFloat(number, 64)
	mare.PanicIfErr(err)

	return parsedNumber
}

func parseFlags() (float64, params) {
	var inputUnit, outputUnit string
	var precisionPoints int

	flag.StringVar(&inputUnit, "i", "B", "input unit")
	flag.StringVar(&outputUnit, "o", "G", "output unit")
	flag.IntVar(&precisionPoints, "p", defaultPrecision, "number of decimal precision digits")

	flag.Parse()

	ensureCorrectNumArgs()
	parsedNumber := parseNumber()

	return parsedNumber, params{inputUnit: inputUnit, out: outputUnit, precision: precisionPoints}
}

func main() {
	quantity, parameters := parseFlags()

	converted := convert(parameters.inputUnit, parameters.out, quantity)
	formatString := getFormatString(parameters.precision)

	fmt.Printf(formatString, converted)
}
