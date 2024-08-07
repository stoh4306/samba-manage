package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("*******************************")
	fmt.Println(" M2M-STORAGE MANAGEMENT S/W")
	fmt.Println("*******************************")

	smb_conf_path := os.Args[1]

	// Read smb.share.conf
	shareFolder, err := read_smb_conf_file(smb_conf_path)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create a share folder array and set
	var sfa ShareFolderArray
	sfa.name = "Test share folder array"
	sfa.shareFolders = shareFolder

	// add a user to valid users
	err = sfa.addToValidUsers("user1", "smbshare")
	if err != nil {
		fmt.Println(err)
	}

	//-----------------------------------------
	// Add or remove user test
	//-----------------------------------------

	// add user2 to invalid users
	err = sfa.addToInValidUsers("user2", "smbshare")
	if err != nil {
		fmt.Println(err)
	}

	// add user2 to valid users. Note user2 should be removed from invalid users
	err = sfa.addToValidUsers("user2", "smbshare")
	if err != nil {
		fmt.Println(err)
	}

	err = sfa.addToValidUsers("stoh", "smbshare")
	if err != nil {
		fmt.Println(err)
	}

	err = sfa.addToInValidUsers("user2", "smbshare")
	if err != nil {
		fmt.Println(err)
	}

	//------------------------------------------------------------
	// Access privilege
	//------------------------------------------------------------
	err = sfa.addToReadList("user1", "smbshare")
	if err != nil {
		fmt.Println(err)
	}

	err = sfa.addToWriteList("user2", "smbshare")
	if err != nil {
		fmt.Println(err)
	}

	err = sfa.addToWriteList("user1", "smbshare")
	if err != nil {
		fmt.Println(err)
	}

	err = sfa.printShareFolderData()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = sfa.exportShareFolderData("./smb.share.conf")
	if err != nil {
		fmt.Println(err)
		return
	}
}
