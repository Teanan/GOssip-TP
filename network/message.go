package network

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Message struct {
	Kind string
	Data string
}

type formatError struct {
	raw string
}

func GetNextMessage(conn net.Conn) (Message, error) {
	data, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil {
		return Message{}, err
	}

	return createMessage(strings.TrimSpace(data))
}

func createMessage(raw string) (Message, error) {
	rawSplit := strings.SplitN(raw, " ", 2)

	if len(rawSplit) < 1 {
		return Message{}, &formatError{raw}
	}

	if len(rawSplit) == 2 {
		return Message{rawSplit[0], rawSplit[1]}, nil
	} else {
		return Message{rawSplit[0], ""}, nil
	}
}

func (m Message) Send(conn net.Conn) error {
	fmt.Println("Sent :", m)
	_, err := conn.Write([]byte(m.Kind + " " + strings.TrimSpace(m.Data) + "\n"))
	return err
}

func (m Message) String() string {
	if m.Data != "" {
		return fmt.Sprintf("[%s] %s", m.Kind, m.Data)
	} else {
		return fmt.Sprintf("[%s] %s", m.Kind, "<nil>")
	}
}

func (e *formatError) Error() string {
	return fmt.Sprintf("Malformated message : %s", e.raw)
}
