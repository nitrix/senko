package app

import (
	"bytes"
	"encoding/gob"
	"strings"
)

type User struct {
	gatewayName string
	identifier string
}

func init() {
	gob.Register(User{})
}

func (u User) GobEncode() ([]byte, error) {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(u.String())
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (u *User) GobDecode(data []byte) error {
	var str string

	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(&str)
	if err != nil {
		return err
	}

	parts := strings.Split(str, "/")
	u.gatewayName = parts[0]
	u.identifier = parts[1]

	return nil
}

func (u User) String() string {
	return u.gatewayName + "/" + u.identifier
}