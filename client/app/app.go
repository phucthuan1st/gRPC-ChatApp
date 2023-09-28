package app

import (
	"context"
	"fmt"
	"io"

	gs "github.com/phucthuan1st/gRPC-ChatRoom/grpcService"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientApp struct {
	app        *tview.Application
	username   *string
	serverAddr *string
	client     gs.ChatRoomClient
	conn       *grpc.ClientConn
	nav        *tview.Pages
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
	stream, _ := ca.client.Listen(context.Background(), &gs.Command{AdditionalInfo: ca.username})

	// start to wait for server message
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}

			// TODO: update code to handle incoming message
			if err != nil {
				fmt.Printf("Failed to receive a note : %v\n", err) // log
			}
			fmt.Printf("%s just chat: %s\n", in.GetSender(), in.GetMessage())
		}
	}()
	<-waitc
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

	messageView := tview.NewTextView().SetTextAlign(tview.AlignLeft)
	messageFlex := tview.NewFlex().AddItem(messageView, 0, 0, true)
	messageFlex.SetBorder(true).SetTitle("Message")

	inputFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	inputFlex.SetBorder(true)
	inputFlex.AddItem(tview.NewTextArea(), 0, 4, true)
	inputFlex.AddItem(tview.NewButton("Send"), 0, 1, false)

	leftFlex.AddItem(messageView, 0, 9, false)
	leftFlex.AddItem(inputFlex, 0, 1, false)
	leftFlex.SetBorder(true)

	onlineClientView := tview.NewList()
	onlineClientView.SetBorder(true).SetTitle("Online Clients")

	rightFlex.AddItem(onlineClientView, 0, 9, false)
	rightFlex.AddItem(tview.NewButton("Logout"), 0, 1, false)
	rightFlex.SetBorder(true)

	flex.AddItem(leftFlex, 0, 3, true)
	flex.AddItem(rightFlex, 0, 1, false)

	ca.nav.AddAndSwitchToPage("ChatPage", flex, true)
}

func updateMessageTextView(textView *tview.TextView, message string) {
	currentText := textView.GetText(false)
	if currentText != "" {
		currentText += "\n"
	}
	textView.SetText(fmt.Sprintf("%sYou: %s", currentText, message))
}
