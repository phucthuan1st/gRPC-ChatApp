package gui

import (
	"github.com/gotk3/gotk3/gtk"
)

// CreateChatClientGUI creates a chat client GUI and returns a pointer to the window.
func CreateChatClientGUI() *gtk.Window {
	// Initialize GTK
	gtk.Init(nil)

	// Create a new window
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetTitle("Chat Client")
	window.SetDefaultSize(400, 300)

	// Create a chat display area
	textView := gtk.NewTextView()
	textView.SetEditable(false)
	textView.SetWrapMode(gtk.WRAP_WORD)

	// Create an input field
	entry := gtk.NewEntry()
	entry.SetEditable(true)

	// Create a send button
	sendButton := gtk.NewButtonWithLabel("Send")

	// Create a scrolled window to contain the chat display area
	scrolledWindow := gtk.NewScrolledWindow(nil, nil)
	scrolledWindow.Add(textView)

	// Create a vertical box to arrange widgets
	vbox := gtk.NewVBox(false, 0)
	vbox.PackStart(scrolledWindow, true, true, 0)
	vbox.PackStart(entry, false, false, 0)
	vbox.PackStart(sendButton, false, false, 0)

	// Add the vertical box to the main window
	window.Add(vbox)

	// Show all widgets
	window.ShowAll()

	// Handle window close event
	window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	return window
}
