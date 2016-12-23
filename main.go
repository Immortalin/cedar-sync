package main

import (
	"log"
	"mime"
	"os"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func main() {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	pw := os.Getenv("gmail_pw")
	username := os.Getenv("gmail_username")
	if pw == "" {
		log.Fatal("Password Environmental Variable Empty!")
	}
	log.Println(pw)
	if err := c.Login(username, pw); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	go func() {
		// c.List will send mailboxes to the channel and close it when done
		if err := c.List("", "*", mailboxes); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := &imap.SeqSet{}
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	go func() {
		if err := c.Fetch(seqset, []string{imap.EnvelopeMsgAttr}, messages); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Last 4 messages:")
	dec := new(mime.WordDecoder)
	for msg := range messages {
		if subject, err := dec.DecodeHeader(msg.Envelope.Subject); err == nil {
			log.Println("* " + subject)
		}
	}

	log.Println("Done!")
}
