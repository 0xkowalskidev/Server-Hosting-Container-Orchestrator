package models

import "encoding/json"

type Namespace struct {
	ID string // Id is namespace value
}

func (n Namespace) Key() string {
	return "/namespaces/" + n.ID
}

func (n Namespace) Value() (string, error) {
	bytes, err := json.Marshal(n)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
