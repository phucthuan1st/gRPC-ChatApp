package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	gs "github.com/phucthuan1st/gRPC-ChatRoom/grpcService"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

// Client App for gRPC-ChatRoom service usage.
type ClientApp struct {
	app                 *tview.Application
	username            *string
	conn                *grpc.ClientConn
	stub                gs.ChatRoomClient
	chatStream          gs.ChatRoom_ChatClient
	navigator           *tview.Pages
	publicMessageList   *tview.List
	privateMessageList  map[string]*tview.List
	connectedClientList *tview.List
	inputArea           *tview.TextArea
	selectedIndex       int
	stillRunning        bool
	refreshFuncs        []func()
	RefreshInterval     time.Duration
	Port                int
	Ipaddr              string
	nRecieveMessage     int
}

// Start and run the client application
func (ca *ClientApp) Start() error {
	var err error

	serverAddr := fmt.Sprintf("%s:%d", ca.Ipaddr, ca.Port)
	ca.conn, err = grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	ca.app = tview.NewApplication()

	if err != nil {
		ca.alert("Cannot connect to server", "")
		ca.Exit()
	} else {
		ca.stub = gs.NewChatRoomClient(ca.conn)
	}

	ca.connectedClientList = tview.NewList()
	ca.publicMessageList = tview.NewList()
	ca.privateMessageList = make(map[string]*tview.List)
	ca.navigator = tview.NewPages()

	ca.inputArea = tview.NewTextArea()

	ca.refreshFuncs = append(ca.refreshFuncs, func() {
		ca.app.Draw()
	})

	ca.nRecieveMessage = 0

	ca.navigateToLogin()
	ca.app.SetRoot(ca.navigator, true).EnableMouse(true).Run()

	return err
}

// Start listening for messages from server
func (ca *ClientApp) startListening() {
	stream, err := ca.stub.Chat(context.Background())
	if err != nil {
		ca.alert("Cannot connect to server", "")
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
					ca.Exit()
					ca.alert("Disconnected from server", "Login")
					return
				} else {
					if msg.GetPrivate() > 0 {
						ca.updatePrivateMessageList(msg.GetSender(), msg.GetSender(), msg.GetMessage())
						ca.nRecieveMessage++
					} else {
						ca.updateMessageList(msg.GetSender(), msg.GetMessage())
					}
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
		ca.alert("Failed to request authentication from server. Please try again", "Login")
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

	ca.alert("Unexpected error", "Login")
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
func (ca *ClientApp) alert(msg string, parentPage string) {

	modal := tview.NewModal().SetText(msg).AddButtons([]string{"Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Cancel" {
				ca.navigator.RemovePage("Alert")
			}
		})

	ca.navigator.AddAndSwitchToPage("Alert", ca.modal(modal, 40, 20), false)

	if parentPage != "" {
		ca.navigator.SwitchToPage(parentPage)
	}
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
					ca.stillRunning = true
					ca.alert("Login successfully!", "")
					go ca.startListening()

					go func() {
						ca.refresh()
						if !ca.stillRunning {
							ca.refreshFuncs = []func(){}
						}
					}()
					ca.navigateToPublicChatRoom()
				}
			} else {
				msg := "Cannot access the input. Please try again later."
				ca.alert(msg, "Login")
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
				ca.alert(err.Error(), "Register")
			} else {
				ca.alert(fmt.Sprintf("User %s registered successfully with code %d!", result.GetUsername(), result.GetStatus()), "")
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
	if ca.navigator.HasPage("Login") {
		ca.navigator.SwitchToPage("Login")
	} else {
		flex := ca.createCenterFlexForm(ca.createLoginForm(), false)
		ca.navigator.AddAndSwitchToPage("Login", flex, true)
	}
}

// Register page navigation
func (ca *ClientApp) navigateToRegister() {
	if ca.navigator.HasPage("Register") {
		ca.navigator.SwitchToPage("Register")
	} else {
		flex := ca.createCenterFlexForm(ca.createUserRegistrationForm(), true)
		ca.navigator.AddAndSwitchToPage("Register", flex, true)
	}
}

// Quit the application
func (ca *ClientApp) Exit() {
	ca.stillRunning = false
	ca.refreshFuncs = []func(){}
	ca.conn.Close()
	ca.app.Stop()
}

// left area of chat room layout
func (ca *ClientApp) createChatRoomLeftFlex() *tview.Flex {
	leftFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Message view displays the chat room messages from both current user and other users
	ca.publicMessageList.SetBorder(true).SetTitle("Messages").SetTitleAlign(tview.AlignRight)

	// Input flex contains the input field and the send button
	ca.inputArea.SetText("", true)

	sendBtn := tview.NewButton("Send")
	sendBtn.SetSelectedFunc(func() {
		message := ca.inputArea.GetText()

		if message != "" {
			ca.updateMessageList("You", message)

			ca.chatStream.Send(&gs.ChatMessage{
				Sender:  *ca.username,
				Message: message,
			})

			ca.inputArea.SetText("", true)
		}
	})

	inputFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	inputFlex.SetBorder(true)
	inputFlex.SetTitleAlign(tview.AlignLeft)
	inputFlex.SetTitle(*ca.username)
	inputFlex.AddItem(ca.inputArea, 0, 5, true)
	inputFlex.AddItem(sendBtn, 0, 1, false)

	// Add the message flex and the input flex to the left flex
	leftFlex.AddItem(ca.publicMessageList, 0, 7, true)
	leftFlex.AddItem(inputFlex, 0, 2, false)
	leftFlex.SetBorder(true)

	return leftFlex
}

// right area of chat room layout
func (ca *ClientApp) createChatRoomRightFlex() *tview.Flex {
	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	quitBtn := tview.NewButton("Quit")
	quitBtn.SetBorder(true)
	quitBtn.SetSelectedFunc(func() {
		ca.Exit()
	})

	ca.connectedClientList.SetBorder(true).SetTitle("Online Clients")

	logoutBtn := tview.NewButton("Logout")
	logoutBtn.SetBorder(true)
	logoutBtn.SetSelectedFunc(func() {
		ca.Exit()
		ca.Start()
	})

	rightFlex.AddItem(quitBtn, 0, 1, false)
	rightFlex.AddItem(ca.connectedClientList, 0, 9, false)
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
	if ca.navigator.HasPage("Public Chat Room") {
		ca.navigator.SwitchToPage("Public Chat Room")
	} else {
		flex := ca.CreateChatRoom()
		ca.refreshFuncs = append(ca.refreshFuncs, func() {
			mu := sync.Mutex{}

			mu.Lock()
			ca.selectedIndex = ca.connectedClientList.GetCurrentItem()
			mu.Unlock()

			ca.updateOnlineClientsList()
		})
		ca.navigator.AddAndSwitchToPage("Public Chat Room", flex, true)
	}
}

// update the message text view with the new incoming message
func (ca *ClientApp) updateMessageList(sender, message string) {
	var r rune
	if sender == "You" {
		r = '>'
	} else if sender == "Server" {
		r = 'o'
	} else {
		r = '<'
	}

	likeHandler := func() {
		if sender == "You" || sender == "Server" {
			return
		}

		ca.stub.LikeMessage(context.Background(), &gs.UserRequest{
			Sender: *ca.username,
			Target: &sender,
		})

	}

	ca.publicMessageList.SetCurrentItem(ca.nRecieveMessage)
	ca.publicMessageList.AddItem(sender, message, r, likeHandler)
	ca.app.SetFocus(ca.publicMessageList)
}

// update the message text view with the new incoming message
func (ca *ClientApp) updatePrivateMessageList(sender, target, message string) {
	var r rune
	if sender == "You" {
		r = '>'
	} else if sender == "Server" {
		r = 'o'
	} else {
		r = '<'
	}

	if _, ok := ca.privateMessageList[target]; !ok {
		ca.privateMessageList[target] = tview.NewList()
	}

	ca.privateMessageList[target].AddItem(sender, message, r, nil)
	ca.app.SetFocus(ca.privateMessageList[target])
}

// update the connected clients list
func (ca *ClientApp) updateOnlineClientsList() {
	// Create a UserRequest with the current user's username as the sender
	userRequest := &gs.UserRequest{
		Sender: *ca.username,
	}

	// Get the list of connected clients from the server
	connectedClients, err := ca.stub.GetConnectedPeers(context.Background(), userRequest)
	if err != nil {
		ca.alert(fmt.Sprintf("Failed to get connected clients: %v", err), "")
		return
	}

	// Clear the current list and update with the new list
	ca.connectedClientList.Clear()

	// Add each connected client to the list
	for index, username := range connectedClients.GetUsername() {
		status := connectedClients.Status[index]

		if status == "Offline" {
			ca.connectedClientList.AddItem(username, status, 'x', nil)
		} else {
			ca.connectedClientList.AddItem(username, status, '+', ca.startAPrivateSession)
		}
	}
	ca.connectedClientList.SetCurrentItem(ca.selectedIndex)

	if ca.app.GetFocus() != ca.inputArea {
		ca.app.SetFocus(ca.connectedClientList)
	}
}

func (ca *ClientApp) startAPrivateSession() {
	target, status := ca.connectedClientList.GetItemText(ca.selectedIndex)

	if status != "Offline" {
		ca.navigateToPrivateChatRoom(target)
	}
}

func (ca *ClientApp) createUserProfieView(target string) *tview.TextView {
	userProfView := tview.NewTextView()
	userProfView.SetBorder(true).SetTitle("Profile")
	userProfView.SetTitleAlign(tview.AlignRight)

	profile, err := ca.stub.GetPeerInfomations(context.Background(), &gs.UserRequest{
		Target: &target,
		Sender: *ca.username,
	})

	if err != nil {
		ca.alert(fmt.Sprintf("Failed to get user information: %v", err), "")
		return nil
	}
	profileString := fmt.Sprintf("Full Name: %s\nUsername: %s\nEmail: %s\nBirthday: %s\nAddress: %v", profile.GetFullName(), profile.GetUsername(), profile.GetEmail(), profile.GetBirthdate(), profile.GetAddress())
	userProfView.SetText(profileString)

	return userProfView
}

// left area of private chat room layout
func (ca *ClientApp) createPrivateChatRoomLeftFlex(target string) *tview.Flex {
	leftFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// COMPONENT: Message view displays the chat room messages from both current user and other users
	if _, ok := ca.privateMessageList[target]; !ok {
		ca.privateMessageList[target] = tview.NewList()
	}
	ca.privateMessageList[target].SetBorder(true).SetTitle(target).SetTitleAlign(tview.AlignRight)

	// COMPONENT: Input flex contains the input field and the send button
	ca.inputArea.SetText("", true)

	sendBtn := tview.NewButton("Send")
	sendBtn.SetSelectedFunc(func() {
		message := ca.inputArea.GetText()

		if message != "" {
			ca.updatePrivateMessageList("You", target, message)

			ca.stub.SendPrivateMessage(context.Background(), &gs.PrivateChatMessage{
				Sender:   *ca.username,
				Recipent: target,
				Message:  message,
			})

			ca.inputArea.SetText("", true)
		}
	})

	inputFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	inputFlex.SetBorder(true)
	inputFlex.SetTitleAlign(tview.AlignLeft)
	inputFlex.SetTitle(*ca.username)
	inputFlex.AddItem(ca.inputArea, 0, 5, true)
	inputFlex.AddItem(sendBtn, 0, 1, false)

	// COMPONENT: tabbed flex
	tabbedFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	msgTabBtn := tview.NewButton("Messages")
	msgTabBtn.SetBorder(true)
	msgTabBtn.SetDisabled(true)

	profTabBtn := tview.NewButton("Profile")
	profTabBtn.SetBorder(true)

	returnTabBtn := tview.NewButton("Chat Room")
	returnTabBtn.SetBorder(true)
	returnTabBtn.SetSelectedFunc(func() {
		ca.navigateToPublicChatRoom()
	})

	spacer := ca.createSpacer()
	tabbedFlex.AddItem(msgTabBtn, 0, 1, false)
	tabbedFlex.AddItem(profTabBtn, 0, 1, false)
	tabbedFlex.AddItem(spacer, 0, 2, false)
	tabbedFlex.AddItem(returnTabBtn, 0, 2, false)

	// FINAL: Add the message flex and the input flex to the left flex
	leftFlex.AddItem(tabbedFlex, 0, 1, false)
	leftFlex.AddItem(ca.privateMessageList[target], 0, 6, true)
	leftFlex.AddItem(inputFlex, 0, 2, false)
	leftFlex.SetBorder(true)

	profTabBtn.SetSelectedFunc(func() {
		msgTabBtn.SetDisabled(false)
		profTabBtn.SetDisabled(true)
		leftFlex.Clear()

		userProfView := ca.createUserProfieView(target)
		leftFlex.AddItem(tabbedFlex, 0, 1, false)
		leftFlex.AddItem(userProfView, 0, 8, true)
	})

	msgTabBtn.SetSelectedFunc(func() {
		msgTabBtn.SetDisabled(true)
		profTabBtn.SetDisabled(false)
		leftFlex.Clear()

		leftFlex.AddItem(tabbedFlex, 0, 1, false)
		leftFlex.AddItem(ca.privateMessageList[target], 0, 6, true)
		leftFlex.AddItem(inputFlex, 0, 2, false)
	})

	return leftFlex
}

// create a private chat room page in flex
func (ca *ClientApp) createPrivateChatRoom(target string) *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex.SetTitle("Private Chat Room")

	leftFlex := ca.createPrivateChatRoomLeftFlex(target)
	rightFlex := ca.createChatRoomRightFlex()

	flex.AddItem(leftFlex, 0, 3, false)
	flex.AddItem(rightFlex, 0, 1, false)

	return flex
}

// navigate to the private chat room
func (ca *ClientApp) navigateToPrivateChatRoom(target string) {
	if ca.navigator.HasPage("Private Chat Room " + target) {
		ca.navigator.SwitchToPage("Private Chat Room " + target)
	} else {
		flex := ca.createPrivateChatRoom(target)
		ca.navigator.AddAndSwitchToPage("Private Chat Room "+target, flex, true)
	}
}

// refresh app (including refresh connected clients list)
func (ca *ClientApp) refresh() {
	tick := time.NewTicker(ca.RefreshInterval)
	defer tick.Stop() // Ensure the ticker is stopped when the function exits.

	for {
		select {
		case <-tick.C:
			for _, fun := range ca.refreshFuncs {
				fun()
			}
		}
	}
}
