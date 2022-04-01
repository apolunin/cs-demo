package common

import (
  "encoding/json"
  "errors"
)

type Message struct {
  Action *ActionType `json:"action"`
  Key    string      `json:"key"`
  Value  string      `json:"value"`
}

type ActionType string

const (
  AddItem     ActionType = "AddItem"
  GetItem     ActionType = "GetItem"
  RemoveItem  ActionType = "RemoveItem"
  GetAllItems ActionType = "GetAllItems"
)

func (at *ActionType) IsValid() bool {
  switch *at {
  case AddItem,
    GetItem,
    RemoveItem,
    GetAllItems:

    return true
  }

  return false
}

func (at *ActionType) UnmarshallJSON(b []byte) error {
  var s string
  if err := json.Unmarshal(b, &s); err != nil {
    return err
  }
  if actionType := ActionType(s); actionType.IsValid() {
    *at = actionType
    return nil
  }

  return errors.New("invalid action type")
}
