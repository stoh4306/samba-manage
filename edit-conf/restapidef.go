package main

import (
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

var smb_conf_file string = "/etc/samba/smb.share.conf"

type SimpleResponse struct {
	Message string `json:"message"`
}

type CreateRequest struct {
	Domain      string `json:"domain"`
	UserId      string `json:"id"`
	StorageName string `json:"storage_name"`
	Quota       string `json:"quota"`
	RootPath    string `json:"root_path"`
}

func validateCreateRequest(cr CreateRequest) string {
	if cr.Domain == "" {
		return "Domain is missing"
	}
	if cr.UserId == "" {
		return "UserID is missing"
	}
	if cr.StorageName == "" {
		return "Storage name is missing"
	}
	if cr.Quota == "" {
		return "Quota is missing"
	}
	if cr.RootPath == "" {
		return "Root path is missing"
	}

	return "OK"
}

func createSharedStorage(c *gin.Context) {
	var cr CreateRequest

	c.BindJSON(&cr)

	// Check if request is correct
	mesg := validateCreateRequest(cr)
	if mesg != "OK" {
		c.IndentedJSON(http.StatusBadRequest,
			SimpleResponse{mesg})
		return
	}

	// Create a user folder if it does not exist
	userFolderPath := cr.RootPath + "/" + cr.UserId + "_" + cr.Domain
	logger.Info("user folder : " + userFolderPath)

	info, err := os.Stat("/" + userFolderPath) // NOTE : "/" at the front of userFolderPath
	if err != nil {                            // Does not exist and create now
		err = exec.Command("zfs", "create", userFolderPath).Run()
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError,
				SimpleResponse{"Failed to create user folder : " + userFolderPath + ", " + err.Error()})
			return
		}
	} else {
		if !info.IsDir() {
			c.IndentedJSON(http.StatusInternalServerError,
				SimpleResponse{"Failed to create user folder : " + userFolderPath})
			return
		}
	}

	// Create a shared folder
	shareFolderPath := userFolderPath + "/" + cr.StorageName
	_, err = os.Stat("/" + shareFolderPath) // NOTE : "/" at the front of userFolderPath
	if err != nil {                         // Does not exist and create now
		err = exec.Command("zfs", "create", shareFolderPath).Run()
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError,
				SimpleResponse{"Failed to create share folder : " + shareFolderPath + "," + err.Error()})
			return
		}
	} else {
		c.IndentedJSON(http.StatusInternalServerError,
			SimpleResponse{"Failed to create share folder : " + shareFolderPath + ", already exists"})
		return
	}

	// Set quota
	err = exec.Command("zfs", "set", "quota="+cr.Quota, shareFolderPath).Run()
	if err != nil {
		exec.Command("zfs", "destroy", shareFolderPath).Run()

		c.IndentedJSON(http.StatusInternalServerError,
			SimpleResponse{"Failed to set quota : " + cr.StorageName})
		return
	}

	// Edit smb.share.conf
	shareFolder, err := read_smb_conf_file(smb_conf_file)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError,
			SimpleResponse{"Failed to change samba configuration"})
		return
	}

	// Create a share folder array and set
	var sfa ShareFolderArray
	sfa.shareFolders = shareFolder

	c.IndentedJSON(http.StatusOK, SimpleResponse{"Successfully create a shared folder : " + cr.StorageName + "," + cr.Quota})
}

type BasicRequest struct {
	StorageName string `json:"storage_name"`
	Domain      string `json:"domain"`
	UserID      string `json:"id"`
	RootPath    string `json:"root_path"`
}

func validateBasicRequest(r BasicRequest) string {
	if r.StorageName == "" {
		return "Storage name is missing"
	}
	if r.Domain == "" {
		return "Domain is missing"
	}
	if r.UserID == "" {
		return "UserID is missing"
	}
	if r.RootPath == "" {
		return "Root path is missing"
	}

	return "OK"
}

func deleteSharedStorage(c *gin.Context) {
	var r BasicRequest

	c.BindJSON(&r)

	// Check if request is correct
	mesg := validateBasicRequest(r)
	if mesg != "OK" {
		c.IndentedJSON(http.StatusBadRequest,
			SimpleResponse{mesg})
		return
	}

	shareFolderPath := r.RootPath + "/" + r.UserID + "_" + r.Domain + "/" + r.StorageName
	err := exec.Command("zfs", "destroy", shareFolderPath).Run()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError,
			SimpleResponse{"Failed to delete " + r.StorageName + ", " + err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, SimpleResponse{"Successfully delete " + r.StorageName})
}

func setQuota(c *gin.Context) {
	var r CreateRequest

	c.BindJSON(&r)

	mesg := validateCreateRequest(r)
	if mesg != "OK" {
		c.IndentedJSON(http.StatusBadRequest,
			SimpleResponse{mesg})
		return
	}

	// Set quota
	shareFolderPath := r.RootPath + "/" + r.UserId + "_" + r.Domain + "/" + r.StorageName
	err := exec.Command("zfs", "set", "quota="+r.Quota, shareFolderPath).Run()
	if err != nil {
		exec.Command("zfs", "destroy", shareFolderPath).Run()

		c.IndentedJSON(http.StatusInternalServerError,
			SimpleResponse{"Failed to set quota : " + r.StorageName + "," + err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, SimpleResponse{"Successfully set quota : " + r.StorageName + "," + r.Quota})
}

func getQuota(c *gin.Context) {
	var r BasicRequest
	c.BindJSON(&r)

	mesg := validateBasicRequest(r)
	if mesg != "OK" {
		c.IndentedJSON(http.StatusBadRequest,
			SimpleResponse{mesg})
		return
	}

	shareFolderPath := r.RootPath + "/" + r.UserID + "_" + r.Domain + "/" + r.StorageName
	output, err := exec.Command("zfs", "list", shareFolderPath).Output()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, SimpleResponse{"Failed to retrieve quota : " + r.StorageName})
		return
	} else {
		tokens := strings.FieldsFunc(string(output), func(r rune) bool {
			return r == ' ' || r == '\n'
		})

		used, quota := tokens[6], tokens[7]
		c.IndentedJSON(http.StatusOK, SimpleResponse{used + " " + quota})
	}
}

type SetUserRequest struct {
	StorageName string `json:"storage_name"`
	Domain      string `json:"domain"`
	UserID      string `json:"id"`
	RootPath    string `json:"root_path"`
	Privilege   string `json:"privilege"`
}

func validateSetUserRequest(r SetUserRequest) string {
	if r.StorageName == "" {
		return "Storage name is missing"
	}
	if r.Domain == "" {
		return "Domain is missing"
	}
	if r.UserID == "" {
		return "UserID is missing"
	}
	if r.RootPath == "" {
		return "Root path is missing"
	}
	if r.Privilege == "" {
		return "Privilege is missing"
	}

	return "OK"
}

func setUser(c *gin.Context) {
	var r SetUserRequest

	c.BindJSON(&r)

	mesg := validateSetUserRequest(r)
	if mesg != "OK" {
		c.IndentedJSON(http.StatusBadRequest,
			SimpleResponse{mesg})
		return
	}

	read_smb_conf_file(smb_conf_file)
}
