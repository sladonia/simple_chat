package message_domain

type Message struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

func (m Message) String() string {
	return m.Name + ": " + m.Text
}

func (m Message) Serialize() []byte {
	return []byte(m.String())
}
