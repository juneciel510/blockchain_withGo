package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/manifoldco/promptui"
)

var OperationTypes=[]string{					
							"Transfer coins 'a' -> 'b'",
							"Transfer coins 'b' -> 'c'",
							"Transfer coins 'c' -> 'a'",
							"Print-block Chain",
							"Print-balance for all users",
							"Print-block Chain length",
							"Print-current block",
							}


func ClientLoop(users *Users) {
	var user string
	var amount int
	for {
		validate := func(input string) error {
			_, err := strconv.ParseFloat(input, 64)
			if err != nil {
				return errors.New("Invalid number")
			}
			return nil
		}

		fmt.Println("-----------Enter operation type----------------:")
		for i, v := range OperationTypes {
			fmt.Printf("%v for %q\n", i+1, v)
		}

		prompt := promptui.Prompt{
			Label:    "Operation",
			Validate: validate,
		}

		result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		switch result {
		case "1":
			fmt.Println("Enter the amount you want to transfer: ")
			fmt.Scanln(&amount)
			err:=users.Transfer("a","b","a",amount)
			PrintErr(err)
			break
		case "2":
			fmt.Println("Enter the amount you want to transfer: ")
			fmt.Scanln(&amount)
			err:=users.Transfer("b","c","a",amount)
			PrintErr(err)
			break
		case "3":
			fmt.Println("Enter the amount you want to transfer: ")
			fmt.Scanln(&amount)
			err:=users.Transfer("c","a","a",amount)
			PrintErr(err)
			break
		case "4":
			fmt.Println("Enter the name of user to show its blockchain: ")
			fmt.Scanln(&user)	
			fmt.Println(users.UsersMap[user].Blockchain.String()	)
			break
		case "5":
			fmt.Println("User: 'a'. Balance:",users.UsersMap["a"].GetBalance())
			fmt.Println("User: 'b'. Balance:",users.UsersMap["b"].GetBalance())
			fmt.Println("User: 'c'. Balance:",users.UsersMap["c"].GetBalance())
			break
		case "6":
			fmt.Println("User: 'a', Blockchain length:",len(users.UsersMap["a"].Blockchain.blocks))
			fmt.Println("User: 'b', Blockchain length:",len(users.UsersMap["b"].Blockchain.blocks))
			fmt.Println("User: 'c', Blockchain length:",len(users.UsersMap["c"].Blockchain.blocks))
			break
		case "7":
			block:=users.UsersMap["a"].Blockchain.CurrentBlock()
			fmt.Println(block.StringDetail())
			break
		default:
			break
		}
		
	}
}

