package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

var shareFolderArray_ ShareFolderArray

func main() {
	fmt.Println("*******************************")
	fmt.Println(" M2M-STORAGE MANAGEMENT S/W")
	fmt.Println("*******************************")

	//---------------------------------------
	// Comand-line test
	//---------------------------------------
	//smb_conf_path := os.Args[1]
	//test(smb_conf_path)

	// Load smb.share.conf first
	shareFolders, err := read_smb_conf_file(smb_conf_file)
	if err != nil {
		logger.Error("Failed to read samba conf file : " + smb_conf_file)
		return
	}

	shareFolderArray_.shareFolders = shareFolders

	// For serving rest apis
	router := gin.New()
	router.Use(gin.Logger())

	basePath := "/storage/"

	router.POST(basePath+"/create", createSharedStorage)
	router.POST(basePath+"/delete", deleteSharedStorage)
	router.POST(basePath+"/quota/set", setQuota)
	router.POST(basePath+"/quota/get", getQuota)
	router.POST(basePath+"/user/set", setUser)

	router.Run(":8080")
}

func test(smbconfpath string) {
	// Read smb.share.conf
	shareFolder, err := read_smb_conf_file(smbconfpath)
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
