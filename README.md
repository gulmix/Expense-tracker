# Expense-tracker
Solution for the [expense-tracker]([https://roadmap.sh/projects/task-tracker](https://roadmap.sh/projects/expense-tracker))

##How  to use

Clone the repository and run the following command:

```bash
git clone https://github.com/gulmix/Expense-tracker.git
```

Run the following command to build and run the project:

```bash
go build main.go

# To add an expense
go run main.go add --description "Lunch" --amount 20

# To update an expense
go run main.go update 1 --description "Buy milk"

# To delete an expense
go run maing.go delete --id 1

# To view all expenses.
go run main.go list

# To view a summary of all expenses.
go run main.go summary

# To view a summary of expenses for a specific month (of current year).
go run main.go summary --month 8
```
