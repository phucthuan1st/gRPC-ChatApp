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

type ClientApp struct {
	app         *tview.Application
	username    *string
	serverAddr  *string
	client      gs.ChatRoomClient
	conn        *grpc.ClientConn
	nav         *tview.Pages
	messageView *tview.TextView
	stream      gs.ChatRoom_ChatClient
}

func (ca *ClientApp) Init() {

	// backend infomation
	const port = 55555
	const ipaddr = "localhost"

	var err error

	serverAddr := fmt.Sprintf("%s:%d", ipaddr, port)
	ca.serverAddr = &serverAddr

	ca.conn, err = grpc.Dial(*ca.serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	ca.app = tview.NewApplication()
	ca.nav = tview.NewPages()

	if err != nil {
		ca.Alert("Cannot connect to server")
	} else {
		ca.client = gs.NewChatRoomClient(ca.conn)
	}
	ca.Login()
	ca.app.SetRoot(ca.nav, true).EnableMouse(true).Run()
}

func (ca *ClientApp) StartListening() {
	stream, err := ca.client.Chat(context.Background())
	if err != nil {
		ca.Alert("Cannot connect to server")
	} else {
		ca.stream = stream

		ca.stream.Send(&gs.ChatMessage{
			Sender:  *ca.username,
			Message: fmt.Sprintf("Hello server from %s!", *ca.username),
		})

		go func() {
			for {
				msg, err := ca.stream.Recv()
				if err != nil {
					ca.Alert("Disconnected from server")
					ca.Login()
				} else {
					ca.updateMessageTextView(msg.GetSender(), msg.GetMessage())
				}
			}
		}()
	}
}

func (ca *ClientApp) RequestLogin(password string) bool {
	cred := gs.UserLoginCredentials{
		Username: *ca.username,
		Password: password,
	}

	result, err := ca.client.Login(context.Background(), &cred)
	if err != nil {
		ca.Alert("Failed to request authentication from server. Please try again")
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

	ca.Alert("Unexpected error")
	return false
}

func (ca *ClientApp) Modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

func (ca *ClientApp) Alert(msg string) {

	modal := tview.NewModal().SetText(msg).AddButtons([]string{"Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Cancel" {
				ca.nav.RemovePage("Alert")
			}
		})

	ca.nav.AddAndSwitchToPage("Alert", ca.Modal(modal, 40, 20), false)
}

func (ca *ClientApp) Login() {
	form := tview.NewForm()

	form.AddInputField("Username", "", 30, nil, nil).
		AddPasswordField("Password", "", 30, '*', nil).
		AddButton("Login", func() {
			usernameField, usernameFieldFound := form.GetFormItemByLabel("Username").(*tview.InputField)
			passwordField, passwordFieldFound := form.GetFormItemByLabel("Password").(*tview.InputField)

			if usernameFieldFound && passwordFieldFound {
				username := usernameField.GetText()
				ca.username = &username

				password := passwordField.GetText()
				isAuthenticated := ca.RequestLogin(password)
				if !isAuthenticated {
					usernameField.SetText("")
					passwordField.SetText("")
				} else {
					ca.Alert("Login successfully!")
					go ca.StartListening()
					ca.JoinChat()
				}
			} else {
				msg := "Cannot access the input. Please try again later."
				ca.Alert(msg)
			}
		}).
		AddButton("Register", func() {
			ca.nav.SwitchToPage("Register")
		}).
		AddButton("Quit", func() {
			ca.Exit()
		})

	form.SetBorder(true).SetTitle("gRPC Chat Room").SetTitleAlign(tview.AlignLeft)

	// Add an empty text view as a spacer to push the form to the center
	spacer := tview.NewTextView().
		SetText("").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)

	// Set the flex properties to center the form
	flex := tview.NewFlex().
		AddItem(spacer, 0, 1, false). // left spacer
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
						AddItem(spacer, 0, 1, false). // top spacer
						AddItem(form, 0, 1, true).
						AddItem(spacer, 0, 1, false), 0, 1, false). // bottom spacer
		AddItem(spacer, 0, 1, false) // right spacer

	ca.nav.AddAndSwitchToPage("Login", flex, true)
}

func (ca *ClientApp) Exit() {
	ca.conn.Close()
	ca.app.Stop()
}

func (ca *ClientApp) JoinChat() {

	flex := tview.NewFlex().SetDirection(tview.FlexColumn)
	flex.SetTitle("Chat Room")

	leftFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	ca.messageView = tview.NewTextView().SetTextAlign(tview.AlignLeft)
	messageFlex := tview.NewFlex().AddItem(ca.messageView, 0, 0, true)
	messageFlex.SetBorder(true).SetTitle("Message")

	inputFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	inputFlex.SetBorder(true)
	inputArea := tview.NewTextArea()
	inputFlex.AddItem(inputArea, 0, 4, true)
	sendBtn := tview.NewButton("Send")
	sendBtn.SetSelectedFunc(func() {
		message := inputArea.GetText()

		if inputArea.GetText() != "" {
			inputArea.SetText("", true)
			ca.updateMessageTextView("You", message)

			ca.stream.Send(&gs.ChatMessage{
				Sender:  *ca.username,
				Message: message,
			})
		}
	})
	inputFlex.AddItem(sendBtn, 0, 1, false)

	leftFlex.AddItem(ca.messageView, 0, 9, false)
	leftFlex.AddItem(inputFlex, 0, 1, false)
	leftFlex.SetBorder(true)

	onlineClientView := tview.NewList()
	onlineClientView.SetBorder(true).SetTitle("Online Clients")

	rightFlex.AddItem(onlineClientView, 0, 9, false)
	logoutBtn := tview.NewButton("Logout")
	logoutBtn.SetSelectedFunc(func() {
		ca.stream.CloseSend()
		ca.Exit()
		ca.Init()
	})
	rightFlex.AddItem(logoutBtn, 0, 1, false)
	rightFlex.SetBorder(true)

	flex.AddItem(leftFlex, 0, 3, true)
	flex.AddItem(rightFlex, 0, 1, false)

	ca.nav.AddAndSwitchToPage("ChatPage", flex, true)
}

func (ca *ClientApp) updateMessageTextView(sender, message string) {
	currentText := ca.messageView.GetText(false)
	if currentText != "" {
		currentText += "\n"
	}
	ca.messageView.SetText(fmt.Sprintf("%s%s: %s", currentText, sender, message))
}
