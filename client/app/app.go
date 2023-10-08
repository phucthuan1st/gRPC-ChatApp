package app

import (
	"context"
	"fmt"

	gs "github.com/phucthuan1st/gRPC-ChatRoom/grpcService"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

// Client App for gRPC-ChatRoom service usage.
type ClientApp struct {
	app         *tview.Application
	username    *string
	serverAddr  *string
	stub        gs.ChatRoomClient
	conn        *grpc.ClientConn
	navigator   *tview.Pages
	messageView *tview.TextView
	chatStream  gs.ChatRoom_ChatClient
}

// Start and run the client application
func (ca *ClientApp) Start() {

	// backend infomation
	const port = 55555
	const ipaddr = "localhost"

	var err error

	serverAddr := fmt.Sprintf("%s:%d", ipaddr, port)
	ca.serverAddr = &serverAddr

	ca.conn, err = grpc.Dial(*ca.serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	ca.app = tview.NewApplication()
	ca.navigator = tview.NewPages()

	if err != nil {
		ca.alert("Cannot connect to server")
	} else {
		ca.stub = gs.NewChatRoomClient(ca.conn)
	}
	ca.navigateToLogin()
	ca.app.SetRoot(ca.navigator, true).EnableMouse(true).Run()
}

// Start listening for messages from server
func (ca *ClientApp) startListening() {
	stream, err := ca.stub.Chat(context.Background())
	if err != nil {
		ca.alert("Cannot connect to server")
	} else {
		ca.chatStream = stream

		ca.chatStream.Send(&gs.ChatMessage{
			Sender:  *ca.username,
			Message: fmt.Sprintf("Hello server from %s!", *ca.username),
		})

		go func() {
			for {
				msg, err := ca.chatStream.Recv()
				if err != nil {
					ca.alert("Disconnected from server")
					return
				} else {
					ca.updateMessageTextView(msg.GetSender(), msg.GetMessage())
				}
			}
		}()
	}
}

// Request for login authentication from the server
func (ca *ClientApp) requestLogin(password string) bool {
	cred := gs.UserLoginCredentials{
		Username: *ca.username,
		Password: password,
	}

	result, err := ca.stub.Login(context.Background(), &cred)
	if err != nil {
		ca.alert("Failed to request authentication from server. Please try again")
		return false
	}

	switch result.Status {
	case int32(codes.Unavailable), int32(codes.Unauthenticated):
		{
			return false
		}
	case int32(codes.OK):
		{
			return true
		}
	}

	ca.alert("Unexpected error")
	return false
}

// An alert modal
func (ca *ClientApp) modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

// Alert a message to the center of the screen
func (ca *ClientApp) alert(msg string) {

	modal := tview.NewModal().SetText(msg).AddButtons([]string{"Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Cancel" {
				ca.navigator.RemovePage("Alert")
			}
		})

	ca.navigator.AddAndSwitchToPage("Alert", ca.modal(modal, 40, 20), false)
}

// a standart login form
func (ca *ClientApp) createLoginForm() *tview.Form {
	form := tview.NewForm()

	form.AddInputField("Username", "", 30, nil, nil).
		AddPasswordField("Password", "", 30, '*', nil).
		AddButton("Login", func() {
			// Retrieve values from the form fields
			usernameField, usernameFieldFound := form.GetFormItemByLabel("Username").(*tview.InputField)
			passwordField, passwordFieldFound := form.GetFormItemByLabel("Password").(*tview.InputField)

			if usernameFieldFound && passwordFieldFound {
				username := usernameField.GetText()
				ca.username = &username

				password := passwordField.GetText()
				isAuthenticated := ca.requestLogin(password)
				if !isAuthenticated {
					usernameField.SetText("")
					passwordField.SetText("")
				} else {
					ca.alert("Login successfully!")
					go ca.startListening()
					ca.navigateToPublicChatRoom()
				}
			} else {
				msg := "Cannot access the input. Please try again later."
				ca.alert(msg)
			}
		}).
		AddButton("Move to Register", func() {
			ca.navigateToRegister()
		}).
		AddButton("Quit", func() {
			ca.Exit()
		})

	form.SetBorder(true).SetTitle("gRPC Chat Room").SetTitleAlign(tview.AlignLeft)

	return form
}

// a standart registration form
func (ca *ClientApp) createUserRegistrationForm() *tview.Form {

	form := tview.NewForm()

	form.AddInputField("Username", "", 30, nil, nil).
		AddPasswordField("Password", "", 30, '*', nil).
		AddInputField("Full Name", "", 30, nil, nil).
		AddInputField("Email (optional)", "", 30, nil, nil).
		AddInputField("Birthdate (optional)", "", 30, nil, nil).
		AddInputField("Street (optional)", "", 30, nil, nil).
		AddInputField("City (optional)", "", 30, nil, nil).
		AddInputField("Country", "", 30, nil, nil).
		AddButton("Register", func() {
			// Retrieve values from the form fields
			usernameField, _ := form.GetFormItemByLabel("Username").(*tview.InputField)
			passwordField, _ := form.GetFormItemByLabel("Password").(*tview.InputField)
			fullNameField, _ := form.GetFormItemByLabel("Full Name").(*tview.InputField)
			emailField, _ := form.GetFormItemByLabel("Email (optional)").(*tview.InputField)
			birthdateField, _ := form.GetFormItemByLabel("Birthdate (optional)").(*tview.InputField)
			streetField, _ := form.GetFormItemByLabel("Street (optional)").(*tview.InputField)
			cityField, _ := form.GetFormItemByLabel("City (optional)").(*tview.InputField)
			countryField, _ := form.GetFormItemByLabel("Country").(*tview.InputField)

			// Create a User message based on the form values
			user := &gs.User{
				Username: usernameField.GetText(),
				Password: passwordField.GetText(),
				FullName: fullNameField.GetText(),
			}

			// Set optional fields if they are not empty
			if emailField.GetText() != "" {
				email := emailField.GetText()
				user.Email = &email
			}
			if birthdateField.GetText() != "" {
				birthdate := birthdateField.GetText()
				user.Birthdate = &birthdate
			}
			if streetField.GetText() != "" || cityField.GetText() != "" || countryField.GetText() != "" {
				street := streetField.GetText()
				city := cityField.GetText()
				user.Address = &gs.Address{
					Street:  &street,
					City:    &city,
					Country: countryField.GetText(),
				}
			}

			// Handle the user registration logic with the created user message
			result, err := ca.stub.Register(context.Background(), user)
			if err != nil {
				ca.alert(err.Error())
			} else {
				ca.alert(fmt.Sprintf("User %s registered successfully with code %d!", result.GetUsername(), result.GetStatus()))
				ca.navigateToLogin()
			}
		}).
		AddButton("Move to Login", func() {
			ca.navigateToLogin()
		}).
		AddButton("Quit", func() {
			ca.Exit()
		})

	form.SetBorder(true).SetTitle("gRPC Chat Room").SetTitleAlign(tview.AlignLeft)

	return form
}

// spacer for flex layout
func (ca *ClientApp) createSpacer() *tview.TextView {
	// Add an empty text view as a spacer to push the form to the center
	spacer := tview.NewTextView().
		SetText("").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	return spacer
}

// a flex layout form with a login form and a register form in the center
func (ca *ClientApp) createCenterFlexForm(form *tview.Form, long_form bool) *tview.Flex {
	// Add an empty text view as a spacer to push the form to the center
	spacer := ca.createSpacer()

	mid_flex := tview.NewFlex().SetDirection(tview.FlexRow)

	if long_form {
		mid_flex.
			AddItem(spacer, 0, 1, false). // top spacer
			AddItem(form, 0, 4, true).
			AddItem(spacer, 0, 1, false) // bottom spacer
	} else {
		mid_flex.
			AddItem(spacer, 0, 1, false). // top spacer
			AddItem(form, 0, 1, true).
			AddItem(spacer, 0, 1, false) // bottom spacer
	}

	// Set the flex properties to center the form
	flex := tview.NewFlex().
		AddItem(spacer, 0, 1, false).   // left spacer
		AddItem(mid_flex, 0, 2, false). // bottom spacer
		AddItem(spacer, 0, 1, false)    // right spacer

	return flex
}

// Login page navigation
func (ca *ClientApp) navigateToLogin() {
	flex := ca.createCenterFlexForm(ca.createLoginForm(), false)
	ca.navigator.AddAndSwitchToPage("Login", flex, true)
}

// Register page navigation
func (ca *ClientApp) navigateToRegister() {
	flex := ca.createCenterFlexForm(ca.createUserRegistrationForm(), true)
	ca.navigator.AddAndSwitchToPage("Register", flex, true)
}

// Quit the application
func (ca *ClientApp) Exit() {
	ca.conn.Close()
	ca.app.Stop()
}

// left area of chat room layout
func (ca *ClientApp) createChatRoomLeftFlex() *tview.Flex {
	leftFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Message view displays the chat room messages from both current user and other users
	ca.messageView = tview.NewTextView().SetTextAlign(tview.AlignLeft)
	ca.messageView.SetBorder(true).SetTitle("Messages").SetTitleAlign(tview.AlignRight)

	// Input flex contains the input field and the send button
	inputArea := tview.NewTextArea()
	inputArea.SetBorder(true)

	sendBtn := tview.NewButton("Send")
	sendBtn.SetBorder(true)
	sendBtn.SetSelectedFunc(func() {
		message := inputArea.GetText()

		if inputArea.GetText() != "" {
			inputArea.SetText("", true)
			ca.updateMessageTextView("You", message)

			ca.chatStream.Send(&gs.ChatMessage{
				Sender:  *ca.username,
				Message: message,
			})
		}
	})

	inputFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	inputFlex.AddItem(inputArea, 0, 5, true)
	inputFlex.AddItem(sendBtn, 0, 1, false)

	// Add the message flex and the input flex to the left flex
	leftFlex.AddItem(ca.messageView, 0, 9, true)
	leftFlex.AddItem(inputFlex, 0, 1, false)
	leftFlex.SetBorder(true)

	return leftFlex
}

// right area of chat room layout
func (ca *ClientApp) createChatRoomRightFlex() *tview.Flex {
	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	onlineClientView := tview.NewList()
	onlineClientView.SetBorder(true).SetTitle("Online Clients")

	rightFlex.AddItem(onlineClientView, 0, 9, false)
	logoutBtn := tview.NewButton("Logout")
	logoutBtn.SetSelectedFunc(func() {
		ca.Exit()
		ca.Start()
	})
	rightFlex.AddItem(logoutBtn, 0, 1, false)
	rightFlex.SetBorder(true)

	return rightFlex
}

// create a chat room page in flex
func (ca *ClientApp) CreateChatRoom() *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex.SetTitle("Chat Room")

	leftFlex := ca.createChatRoomLeftFlex()
	rightFlex := ca.createChatRoomRightFlex()

	flex.AddItem(leftFlex, 0, 3, false)
	flex.AddItem(rightFlex, 0, 1, false)

	return flex
}

// navigate to the public chat room page
func (ca *ClientApp) navigateToPublicChatRoom() {
	flex := ca.CreateChatRoom()
	ca.navigator.AddAndSwitchToPage("ChatPage", flex, true)
}

// update the message text view with the new incomming message
func (ca *ClientApp) updateMessageTextView(sender, message string) {
	currentText := ca.messageView.GetText(false)
	if currentText != "" {
		currentText += "\n"
	}
	ca.messageView.SetText(fmt.Sprintf("%s%s: %s", currentText, sender, message))
}
