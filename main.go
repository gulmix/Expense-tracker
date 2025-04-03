package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

type Expense struct {
	ID          int       `json:"id"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
}

func main() {
	var rootCmd = &cobra.Command{Use: "expense-tracker"}
	rootCmd.AddCommand(addCmd, deleteCmd, updateCmd, listCmd, summaryCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new expense",
	Run: func(cmd *cobra.Command, args []string) {
		description, _ := cmd.Flags().GetString("description")
		amount, _ := cmd.Flags().GetFloat64("amount")

		if amount <= 0 {
			fmt.Println("Error: amount must be a positive number")
			return
		}

		expenses, err := loadExpenses()
		if err != nil {
			fmt.Println("Error loading expenses:", err)
			return
		}

		newID := 0
		for _, e := range expenses {
			if e.ID > newID {
				newID = e.ID
			}
		}
		newID++

		newExpense := Expense{
			ID:          newID,
			Date:        time.Now(),
			Description: description,
			Amount:      amount,
		}

		expenses = append(expenses, newExpense)

		err = saveExpenses(expenses)
		if err != nil {
			fmt.Println("Error saving expenses:", err)
			return
		}

		fmt.Printf("Expense added successfully (ID: %d)\n", newID)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an expense by ID",
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetInt("id")

		expenses, err := loadExpenses()
		if err != nil {
			fmt.Println("Error loading expenses:", err)
			return
		}

		found := false
		for i, e := range expenses {
			if e.ID == id {
				expenses = append(expenses[:i], expenses[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("Expense with ID %d not found\n", id)
			return
		}

		err = saveExpenses(expenses)
		if err != nil {
			fmt.Println("Error saving expenses:", err)
			return
		}

		fmt.Println("Expense deleted successfully")
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing expense",
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetInt("id")

		expenses, err := loadExpenses()
		if err != nil {
			fmt.Println("Error loading expenses:", err)
			return
		}

		found := -1
		for i, e := range expenses {
			if e.ID == id {
				found = i
				break
			}
		}

		if found == -1 {
			fmt.Printf("Expense with ID %d not found\n", id)
			return
		}

		if !cmd.Flags().Changed("description") && !cmd.Flags().Changed("amount") && !cmd.Flags().Changed("date") {
			fmt.Println("Error: no fields provided to update")
			return
		}

		if cmd.Flags().Changed("description") {
			desc, _ := cmd.Flags().GetString("description")
			expenses[found].Description = desc
		}

		if cmd.Flags().Changed("amount") {
			amount, _ := cmd.Flags().GetFloat64("amount")
			if amount <= 0 {
				fmt.Println("Error: amount must be a positive number")
				return
			}
			expenses[found].Amount = amount
		}

		if cmd.Flags().Changed("date") {
			dateStr, _ := cmd.Flags().GetString("date")
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				fmt.Println("Invalid date format. Use YYYY-MM-DD")
				return
			}
			expenses[found].Date = date
		}

		err = saveExpenses(expenses)
		if err != nil {
			fmt.Println("Error saving expenses:", err)
			return
		}

		fmt.Println("Expense updated successfully")
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all expenses",
	Run: func(cmd *cobra.Command, args []string) {
		expenses, err := loadExpenses()
		if err != nil {
			fmt.Println("Error loading expenses:", err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "# ID\tDate\tDescription\tAmount")
		for _, e := range expenses {
			fmt.Fprintf(w, "# %d\t%s\t%s\t$%.2f\n", e.ID, e.Date.Format("2006-01-02"), e.Description, e.Amount)
		}
		w.Flush()
	},
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show summary of expenses",
	Run: func(cmd *cobra.Command, args []string) {
		expenses, err := loadExpenses()
		if err != nil {
			fmt.Println("Error loading expenses:", err)
			return
		}

		total := 0.0
		currentYear := time.Now().Year()

		if cmd.Flags().Changed("month") {
			month, _ := cmd.Flags().GetInt("month")
			if month < 1 || month > 12 {
				fmt.Println("Invalid month. Must be between 1 and 12.")
				return
			}

			for _, e := range expenses {
				if e.Date.Year() == currentYear && e.Date.Month() == time.Month(month) {
					total += e.Amount
				}
			}

			monthName := time.Month(month).String()
			fmt.Printf("Total expenses for %s: $%.2f\n", monthName, total)
		} else {
			for _, e := range expenses {
				total += e.Amount
			}
			fmt.Printf("Total expenses: $%.2f\n", total)
		}
	},
}

func init() {
	addCmd.Flags().String("description", "", "Description of the expense")
	addCmd.Flags().Float64("amount", 0.0, "Amount of the expense")
	addCmd.MarkFlagRequired("description")
	addCmd.MarkFlagRequired("amount")

	deleteCmd.Flags().Int("id", 0, "ID of the expense to delete")
	deleteCmd.MarkFlagRequired("id")

	updateCmd.Flags().Int("id", 0, "ID of the expense to update")
	updateCmd.Flags().String("description", "", "New description for the expense")
	updateCmd.Flags().Float64("amount", 0.0, "New amount for the expense")
	updateCmd.Flags().String("date", "", "New date for the expense (format: YYYY-MM-DD)")
	updateCmd.MarkFlagRequired("id")

	summaryCmd.Flags().Int("month", 0, "Month to filter expenses (1-12)")
}

func loadExpenses() ([]Expense, error) {
	if _, err := os.Stat("expenses.json"); os.IsNotExist(err) {
		return []Expense{}, nil
	}

	file, err := os.Open("expenses.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var expenses []Expense
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&expenses); err != nil {
		if err.Error() == "EOF" {
			return []Expense{}, nil
		}
		return nil, err
	}

	return expenses, nil
}

func saveExpenses(expenses []Expense) error {
	file, err := os.Create("expenses.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(expenses)
}
