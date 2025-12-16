package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
)

type account struct {
	// All values are in cents.
	value      int64
	growthRate float64
}

func getenvDollars(key string) int64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		log.Fatalf("missing env var: %v", key)
	}
	var val int64
	_, err := fmt.Sscanf(valStr, "%d", &val)
	if err != nil {
		log.Fatalf("unparsable env var: %v = %v", key, valStr)
	}
	return val * 100
}

func getenvFloat(key string) float64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		log.Fatalf("missing env var: %v", key)
	}
	var val float64
	_, err := fmt.Sscanf(valStr, "%f", &val)
	if err != nil {
		log.Fatalf("unparsable env var: %v = %v", key, valStr)
	}
	return val
}

// Requires a .env file with the variables as seen below and gnuplot for plotting.
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	expenses := int64(getenvDollars("expenses"))

	inflationRate := getenvFloat("inflation")
	capGainTax := getenvFloat("capital_gains")

	brokerage := account{value: getenvDollars("brokerage"), growthRate: getenvFloat("brokerage_growth")}
	ira := account{value: getenvDollars("ira"), growthRate: getenvFloat("ira_growth")}

	f, err := os.OpenFile("output.dat", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	w := bufio.NewWriter(f)
	log.SetOutput(w)

	log.SetFlags(0)
	log.Println("Year Brokerage IRA Expenses")
	for year := 2026; year <= 2080; year++ {
		log.Printf("%v %v %v %v\n", year, brokerage.value, ira.value, expenses)
		expenses = int64(float64(expenses) * (1 + inflationRate))
		brokerage.value = int64(float64(brokerage.value) * (1.0 + brokerage.growthRate))
		ira.value = int64(float64(ira.value) * (1.0 + ira.growthRate))
		// This assumes that distributions are made to pay for expenses at the end of the year.
		// TODO: iterate over months, to make this slightly more accurate.
		expensesWithTax := int64(float64(expenses) / (1.0 - capGainTax))
		if brokerage.value > expensesWithTax {
			brokerage.value -= expensesWithTax
		} else {
			if brokerage.value > 0 {
				fmt.Printf("Brokerage exhausted by %d\n", year)
			}
			ira.value += int64(float64(brokerage.value) * (1.0 - capGainTax))
			brokerage.value = 0
			if ira.value > 0 && ira.value <= expenses {
				fmt.Printf("IRA exhausted by %d\n", year)
			}
			ira.value -= expenses
		}
	}

	w.Flush()
	f.Close()

	cmd := exec.Command("C:/Program Files/gnuplot/bin/gnuplot.exe", "output.plt")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("Error getting stdin pipe:", err)
		return
	}
	defer stdin.Close()

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting gnuplot:", err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Println("Error waiting for gnuplot:", err)
	} else {
		fmt.Println("Gnuplot command executed successfully.")
	}
}
