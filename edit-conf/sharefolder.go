package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type ShareFolderArray struct {
	name         string
	shareFolders []ShareFolder
}

func isElement(name string, tokens []string) int {
	index := -1
	for i, token := range tokens {
		if token == name {
			index = i
			break
		}
	}
	return index
}

func addElement(name string, array []string) []string {
	for _, token := range array {
		if token == name {
			return array
		}
	}

	return append(array, name)
}

func removeElement(name string, array []string) []string {
	for i, token := range array {
		if token == name {
			return append(array[:i], array[i+1:]...)
		}
	}
	return array
}

func (sfa *ShareFolderArray) addToReadList(username string, shareName string) error {
	shareId := -1

	for id, shareFolder := range sfa.shareFolders {
		if shareFolder.name == shareName {
			shareId = id
			break
		}
	}

	if shareId < 0 {
		return errors.New("addToReadList(): can't find share folder with name : " + shareName)
	}

	shareFolder := &(sfa.shareFolders)[shareId]
	if isElement(username, shareFolder.valid_users) < 0 || isElement(username, shareFolder.invalid_users) >= 0 {
		return errors.New("addToReadList(): " + username + " is not a valid user")
	}

	shareFolder.read_list = addElement(username, shareFolder.read_list)
	shareFolder.write_list = removeElement(username, shareFolder.write_list)

	return nil
}

func (sfa *ShareFolderArray) addToWriteList(username string, shareName string) error {
	shareId := -1

	for id, shareFolder := range sfa.shareFolders {
		if shareFolder.name == shareName {
			shareId = id
			break
		}
	}

	if shareId < 0 {
		return errors.New("addToReadList(): can't find share folder with name : " + shareName)
	}

	shareFolder := &(sfa.shareFolders)[shareId]
	if isElement(username, shareFolder.valid_users) < 0 || isElement(username, shareFolder.invalid_users) >= 0 {
		return errors.New("addToReadList(): " + username + " is not a valid user")
	}

	shareFolder.write_list = addElement(username, shareFolder.write_list)
	shareFolder.read_list = removeElement(username, shareFolder.read_list)

	return nil
}

func (sfa *ShareFolderArray) addToValidUsers(username string, shareName string) error {
	shareId := -1

	for id, shareFolder := range sfa.shareFolders {
		if shareFolder.name == shareName {
			shareId = id
			break
		}
	}

	if shareId < 0 {
		return errors.New("addToValidUsers(): can't find share folder with name : " + shareName)
	}

	shareFolder := &(sfa.shareFolders)[shareId]
	shareFolder.valid_users = addElement(username, shareFolder.valid_users)
	shareFolder.invalid_users = removeElement(username, shareFolder.invalid_users)

	return nil
}

func (sfa *ShareFolderArray) addToInValidUsers(username string, shareName string) error {
	shareId := -1

	for id, shareFolder := range sfa.shareFolders {
		if shareFolder.name == shareName {
			shareId = id
			break
		}
	}

	if shareId < 0 {
		return errors.New("addToInValidUsers(): can't find share folder with name : " + shareName)
	}

	shareFolder := &(sfa.shareFolders)[shareId]
	shareFolder.invalid_users = addElement(username, shareFolder.invalid_users)
	shareFolder.valid_users = removeElement(username, shareFolder.valid_users)

	return nil
}

func (sfa *ShareFolderArray) exportShareFolderData(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	for _, share := range sfa.shareFolders {
		file.WriteString(share.writeToString())
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

func (sfa *ShareFolderArray) printShareFolderData() error {

	for _, share := range sfa.shareFolders {
		fmt.Print(share.writeToString())
	}

	return nil
}

func removeSpaces(inStr string) string {
	firstValidIndex := 0
	lastValidIndex := len(inStr) - 1

	for i := 0; i < len(inStr); i++ {
		ch := inStr[i]
		if ch != '\t' && ch != ' ' {
			firstValidIndex = i
			break
		}
	}

	for i := 0; i < len(inStr); i++ {
		if inStr[len(inStr)-1-i] != '\t' && inStr[len(inStr)-1-i] != ' ' {
			lastValidIndex = len(inStr) - i
			break
		}
	}

	//fmt.Println([]byte(inStr), firstValidIndex, lastValidIndex)

	if firstValidIndex <= lastValidIndex {
		return inStr[firstValidIndex:lastValidIndex]
	} else {
		return string("")
	}
}

func read_smb_conf_file(smb_conf_path string) ([]ShareFolder, error) {
	var shareFolders []ShareFolder

	file, err := os.Open(smb_conf_path)
	if err != nil {
		return shareFolders, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return shareFolders, err
	}

	lines := strings.Split(string(data), "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if len(line) == 0 || (len(line) >= 1 && (line[0] == '#' || line[0] == ';')) {
			continue
		}
		//fmt.Println(line)

		// Remove comment
		commentSharpStart := strings.IndexRune(line, '#')
		commentSemicolStart := strings.IndexRune(line, ';')
		if commentSharpStart >= 0 && commentSemicolStart >= 0 {

			line = line[0:min(commentSharpStart, commentSemicolStart)]
			//fmt.Println("->", line)
		} else if commentSharpStart >= 0 {
			line = line[0:commentSharpStart]
		} else if commentSemicolStart >= 0 {
			line = line[0:commentSemicolStart]
		}

		if len(line) == 0 {
			continue
		}

		if len(line) > 2 && line[0] == '[' && line[len(line)-1] == ']' {
			shareFolder := ShareFolder{}
			shareFolder.name = removeSpaces(line[1 : len(line)-1])
			shareFolders = append(shareFolders, shareFolder)
			continue
		}

		// Find out the location of the equal sign
		ind_eq := strings.IndexRune(line, '=')
		if ind_eq < 0 {
			continue
		}

		//fmt.Println(line[0:ind_eq], "--->")
		//fmt.Println(line[:ind_eq], line[ind_eq+1:])

		key := removeSpaces(line[:ind_eq])
		value_str := removeSpaces(line[ind_eq+1:])
		// Convert to lower case
		key = strings.ToLower(key)
		//fmt.Println(key)

		currShareFolder := &shareFolders[len(shareFolders)-1]

		switch key {
		case "comment":
			currShareFolder.comment = value_str

		case "path":
			currShareFolder.folder_path = value_str

		case "browsable":
			value := strings.ToLower(value_str)
			if value == "yes" || value == "true" || value_str == "1" {
				currShareFolder.browsable = true
			} else if value == "no" || value == "false" || value_str == "0" {
				currShareFolder.browsable = false
			} else {
				return shareFolders, errors.New("no proper browsable data" + value)
			}

		case "writable":
			value := strings.ToLower(value_str)
			if value == "yes" || value == "true" || value_str == "1" {
				currShareFolder.writable = true
			} else if value == "no" || value == "false" || value_str == "0" {
				currShareFolder.writable = false
			} else {
				return shareFolders, errors.New("no proper writable data : ")
			}

		case "read only":
			value := strings.ToLower(value_str)
			if value == "yes" || value == "true" || value_str == "1" {
				currShareFolder.writable = true
			} else if value == "no" || value == "false" || value_str == "0" {
				currShareFolder.writable = false
			} else {
				return shareFolders, errors.New("no proper writable(read only) data : ")
			}

		case "valid users":
			users := strings.FieldsFunc(value_str, func(r rune) bool {
				return r == ' ' || r == ','
			})
			for _, user := range users {
				currShareFolder.valid_users = append(currShareFolder.valid_users, user)
			}

		case "invalid users":
			users := strings.FieldsFunc(value_str, func(r rune) bool {
				return r == ' ' || r == ','
			})
			for _, user := range users {
				currShareFolder.invalid_users = append(currShareFolder.invalid_users, user)
			}

		case "read list":
			users := strings.FieldsFunc(value_str, func(r rune) bool {
				return r == ' ' || r == ','
			})
			for _, user := range users {
				currShareFolder.read_list = append(currShareFolder.read_list, user)
			}

		case "write list":
			users := strings.FieldsFunc(value_str, func(r rune) bool {
				return r == ' ' || r == ','
			})
			for _, user := range users {
				currShareFolder.write_list = append(currShareFolder.write_list, user)
			}

		case "create mask":
			currShareFolder.create_mask = value_str

		case "directory mask":
			currShareFolder.directory_mask = value_str

		default:
			return shareFolders, errors.New("uknown share folder attribute : " + key)
		}

	}

	err = file.Close()
	if err != nil {
		fmt.Println("Can't close file : ", smb_conf_path)
		return shareFolders, err
	}

	return shareFolders, err
}

type ShareFolder struct {
	name           string
	comment        string
	folder_path    string
	browsable      bool
	writable       bool
	valid_users    []string
	invalid_users  []string
	read_list      []string
	write_list     []string
	create_mask    string
	directory_mask string
}

func (sf ShareFolder) writeToString() string {
	var result string = ""

	result += fmt.Sprintf("[%s]\n", sf.name)
	result += fmt.Sprintf("\t%s = %s\n", "comment", sf.comment)
	result += fmt.Sprintf("\t%s = %s\n", "path", sf.folder_path)

	browsable := "Yes"
	if !sf.browsable {
		browsable = "No"
	}
	result += fmt.Sprintf("\t%s = %s\n", "browsable", browsable)

	writable := "Yes"
	if !sf.writable {
		writable = "No"
	}
	result += fmt.Sprintf("\t%s = %s\n", "writable", writable)

	result += fmt.Sprintf("\t%s = %s\n", "valid users", strings.Join(sf.valid_users, " "))
	result += fmt.Sprintf("\t%s = %s\n", "invalid users", strings.Join(sf.invalid_users, " "))

	result += fmt.Sprintf("\t%s = %s\n", "read list", strings.Join(sf.read_list, " "))
	result += fmt.Sprintf("\t%s = %s\n", "write list", strings.Join(sf.write_list, " "))
	result += fmt.Sprintf("\t%s = %s\n", "create mask", sf.create_mask)
	result += fmt.Sprintf("\t%s = %s\n", "directory mask", sf.directory_mask)

	return result
}
