package app

import (
	"errors"
	"strconv"
	"strings"

	"golang.design/x/hotkey"
)

func getKeys(pressed string) ([]hotkey.Modifier, hotkey.Key, error) {
	keys := strings.Split(pressed, "+")
	if len(keys) == 0 {
		return nil, 0, nil
	}

	var ms []hotkey.Modifier
	var key hotkey.Key
	for _, s := range keys {
		if s == "" {
			panic("empty key")
		}

		switch s {
		case "Ctrl":
			ms = append(ms, hotkey.ModCtrl)
			break
		case "Shift":
			ms = append(ms, hotkey.ModShift)
			break
		case "Alt":
			ms = append(ms, hotkey.ModAlt)
			break
		case "Win":
			ms = append(ms, hotkey.ModWin)
			break
		default:
			if key != 0 {
				return nil, 0, nil
			}

			switch s {
			case "Space":
				key = hotkey.KeySpace
				break
			case "Tab":
				key = hotkey.KeyTab
				break
			case "Enter":
				key = hotkey.KeyReturn
				break
			case "Del":
				key = hotkey.KeyDelete
				break
			case "Esc":
				key = hotkey.KeyEscape
				break
			case "Up":
				key = hotkey.KeyUp
				break
			case "Down":
				key = hotkey.KeyDown
				break
			case "Left":
				key = hotkey.KeyLeft
				break
			case "Right":
				key = hotkey.KeyRight
				break
			}

			if key != 0 {
				break
			}

			if len(s) == 1 {
				key = hotkey.Key(s[0])
			} else if s[0] == 'F' {
				num, err := strconv.Atoi(s[1:])
				if err != nil || num < 1 || num > 12 {
					return nil, 0, nil
				}
				key = hotkey.KeyF1 + hotkey.Key(num-1)
			} else {
				return nil, 0, nil
			}
		}
	}

	if key == 0 {
		panic("key is nil")
	}

	return ms, key, nil
}

func When(press string, action func()) error {
	if press == "" {
		return errors.New("keys is empty")
	}

	if appTask == nil {
		return errors.New("app task is nil")
	}

	ms, key, err := getKeys(press)
	if err != nil {
		return errors.New("invalid keys: " + press)
	}
	go func() {
		hk := hotkey.New(ms, key)

		err = hk.Register()
		if err != nil {
			return
		}

		for range hk.Keydown() {
			action()
		}
	}()
	return nil
}
