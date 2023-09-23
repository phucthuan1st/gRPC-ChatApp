package gui

import (
	"log"

	"github.com/gotk3/gotk3/gtk"
)

// CreateChatClientGUI creates a chat client GUI and returns a pointer to the window.
func CreateChatClientGUI() (gtk.Window, error) {

	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatalf("Failed to create a window view for Client: %s\n", err.Error())
		return gtk.Window{}, err
	}

	window.SetDefaultSize(750, 750)
	window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	return *window, nil
}
