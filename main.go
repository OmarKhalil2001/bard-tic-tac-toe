package main

import (
	"fmt"
	"image/color"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	bard "github.com/aquasecurity/gobard"
)

// Cookies for Bard
var cookie_1 string = "Your __Secure-1PSID Cookie value"
var cookie_2 string = "Your __Secure-1PSIDTS Cookie value "

// Start of GUI
// Defining the Clickable Buttons
var buttons [9]widget.Button = [9]widget.Button{}

// Defining the App Object used to play the game
var a fyne.App = app.New()

// Defining the Second Window of the App to draw the board
var w2 fyne.Window = a.NewWindow("Game")

//==================================================================================
//Defining Helper Functions

// Search for a number in an array
func linearSearch(slice []int, target int) int {
	for i := 0; i < len(slice); i++ {
		if slice[i] == target {
			return i
		}
	}
	return -1
}

// Finds the first digit in a string and returns it
func ParseInt(s string) int {
	for _, i := range s {
		if '0' <= i && i <= '9' {
			return int(i - '0')
		}
	}
	return -1
}

//===========================================================================
//Defining the Game Class

// Define a 2D array of characters to be represent the board
type CharArray [3][3]rune

// The Game struct
type Game struct {
	board CharArray

	//array of currently occupied cells by X
	x_positions []int

	//array of currently occupied cells by O
	o_positions []int
}

// Checks if the Game is over
func (x *Game) isGameOver() bool {
	//Check all horizontal rows first
	for i := range x.board {
		if x.board[i][0] == x.board[i][1] && x.board[i][0] == x.board[i][2] && (x.board[i][0] == 'O' || x.board[i][0] == 'X') {
			return true
		}
	}

	//Checks all the vertical columns
	for i := 0; i < 3; i++ {
		if x.board[0][i] == x.board[1][i] && x.board[2][i] == x.board[0][i] && (x.board[2][i] == 'O' || x.board[2][i] == 'X') {
			return true
		}
	}

	//Check the two diagonals
	if x.board[0][0] == x.board[1][1] && x.board[2][2] == x.board[0][0] && (x.board[0][0] == 'O' || x.board[0][0] == 'X') {
		return true
	}

	if x.board[0][2] == x.board[1][1] && x.board[2][0] == x.board[1][1] && (x.board[1][1] == 'O' || x.board[1][1] == 'X') {
		return true
	}

	return false
}

// Records Player Move
func (x *Game) playerMove(place int) {
	place--
	x.board[place/3][place%3] = 'O'
	x.o_positions = append(x.o_positions, place)
}

// Prepare the prompt for Bard, also returns a list of available cells in case bard fails.
func (x *Game) movePrepare() (string, []int) {

	prompt := "We are currently playing tic tac toe, it is your turn, you play as X. here is how the game looks right now: \n"

	availCells := []int{}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if x.board[i][j] != 'O' && x.board[i][j] != 'X' {
				prompt += strconv.Itoa(i*3 + j)
				availCells = append(availCells, i*3+j)
			} else {
				prompt += string(x.board[i][j])
			}

			if j < 2 {
				prompt += " | "
			}
		}
		prompt += " \n "
	}

	prompt += "You can only play in the cells "

	for _, i := range availCells {
		prompt += strconv.Itoa(i) + ", "
	}

	prompt += "\nReply only with one number, the cell you choose. Do not add anything else to your reply."

	return prompt, availCells
}

// Connects to bard, asks it for a move, records Bard move and in case it fails plays a random move
func (x *Game) GPT_Move() {

	//Connects to Bard using Browser cookies
	client := bard.New(cookie_2, cookie_1)

	//Uses the previous moverPrepare() method to get the prompt and a list of available cells
	prompt, avail := x.movePrepare()

	//First Check if there is any available cells, if no cell is available and no one won yet, then the game is draw
	if len(avail) == 0 {
		fmt.Println("It's a Draw")
		os.Exit(0)
	}

	//sends the prompt to Bard
	err := client.Ask(prompt)
	ans := ""

	if err == nil {
		ans = client.GetAnswer()
	} else {
		fmt.Println(err)
	}

	//Takes the first integer from Bards response, returns -1 if Bard failed
	ansInt := ParseInt(ans)

	//If Bard fails or gives a wrong answer (cell that is not available), play in the first available cell
	if ansInt == -1 || linearSearch(avail, ansInt) == -1 {
		fmt.Println("Bard failed: ", ans)
		ansInt = avail[0]
	}

	//Register Bards Move on the Board
	x.x_positions = append(x.x_positions, ansInt)

	//Disable the Button for the cell Bard chose, so the play can no longer choose to play there
	buttons[ansInt].Disable()
	x.board[ansInt/3][ansInt%3] = 'X'
}

//====================================================================================
//GUI Controllers

// Function called when player clicks the cell he wants to make a move at
func ButtonEvent(i int, buttons *widget.Button, game *Game) func() {
	return func() {
		//First disable the button, so that it can no longer be choosen again.
		buttons.Disable()

		//Record Player Move
		game.playerMove(i + 1)

		//Check if the Player won, if he won, end the game.
		if game.isGameOver() {
			fmt.Println("Congratulations! You won")
			os.Exit(0)
		}

		//Then Bard makes its move
		game.GPT_Move()

		//Check if Bard won. If it won, end the game.
		if game.isGameOver() {
			fmt.Println("Hard Luck! You lost")
			os.Exit(0)
		}

		//Update the Second screen to show the current state of the board
		addSecondScreen(game)
	}
}

// GUI of game controls to the player
func Controller(g *Game) {
	//Creating a Window
	w := a.NewWindow("Bard Tic Tac Toe")

	//Resizing Window
	w.Resize(fyne.NewSize(600, 600))

	id := [9]int{0, 1, 2, 3, 4, 5, 6, 7, 8}

	// Define button size
	buttonSize := fyne.NewSize(200, 200)

	// Creating grid layout to make 3*3 Shape
	grid := container.New(layout.NewGridLayout(3))
	color := color.NRGBA{R: 50, G: 145, B: 89, A: 255}
	rect := canvas.NewRectangle(color)
	//Definig buttons and their handlers
	for i := 0; i < 9; i++ {
		//First, we define buttons, ButtonEvent() gives the proper handler to each button based on its ID
		buttons[i] = *widget.NewButton("", ButtonEvent(id[i], &buttons[i], g))
		//Resizing Buttons
		buttons[i].Resize(buttonSize)

		//Adding Buttons to the Grid Layout
		grid.Add(container.NewStack(&buttons[i], rect))
	}

	//Adding the Content of grid to the screen
	w.SetContent(grid)

	//displaying the first screen
	w.Show()
	//Start the Game
	a.Run()
}

// Function that displays and updates the GUI of board
func addSecondScreen(game *Game) {
	//Creates 9 buttons of the board
	var buttons2 [9]widget.Button = [9]widget.Button{}
	buttonSize := fyne.NewSize(200, 200)

	// Create grid layout
	grid := container.New(layout.NewGridLayout(3))
	k := 0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			//load the x.png and o.png
			xs, err := fyne.LoadResourceFromPath("resources/x.png")
			if err != nil {
				fmt.Println(err)
			}

			os, err := fyne.LoadResourceFromPath("resources/o.png")
			if err != nil {
				fmt.Println(err)
			}

			//assign the proper label to the each cell on the board
			if game.board[i][j] == 'X' {
				buttons2[k] = *widget.NewButtonWithIcon("", xs, func() {})
			} else if game.board[i][j] == 'O' {
				buttons2[k] = *widget.NewButtonWithIcon("", os, func() {})
			} else {
				buttons2[k] = *widget.NewButtonWithIcon("", nil, func() {})
			}
			buttons2[k].Resize(buttonSize)
			grid.Add(&buttons2[k])
			k++
		}
	}

	w2.SetContent(grid)
	w2.Content().Refresh()
}
func main() {
	//Create instance of the game struct and start the game.
	var x Game
	w2.Resize(fyne.NewSize(600, 600))
	w2.Show()
	Controller(&x)
}
