package main

import (
	"os"
	// "fmt"
	"log"
	"encoding/hex"
	"io/ioutil"
	"github.com/codahale/blake2"
	"encoding/json"
)

const filename = "file.txt"
const directory = ".cap"
const refs = "refs"

//Create a "directory" string and a "file" string
//in order to create easy helper functions to string 
//concat everything when generating new files

var commands = map[string]func(){
	"commit": commit,
	"create": create,
	"pull":   pull,
	"push":   push,
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a valid cap command.")
	} else {
		commands[os.Args[1]]()
	}
}

func create() {
	//Create the .cap directory.
	err := os.Mkdir(directory, 0777)
	checkError(err)
	//Create the .refs directory.
	err = os.Mkdir(relativePath(refs), 0777)
	checkError(err)
	//Create an empty "main" branch ref.
	refFile, err := os.Create(".cap/refs/main")
	checkError(err)
	refFile.Close()
}

//Record the blob
//Making a hash of the blob
//Make a commit pointing to the blob (json file)
//Update the ref (hash of commit)
func commit() {
	bytes, err := ioutil.ReadFile(filename)
	checkError(err)
	hash := blake2.NewBlake2B()
	hash.Write(bytes)
	sum := hash.Sum(nil)
	hexValue := hex.EncodeToString(sum)
	//Determine when to use the hex value and when to use the actual hash encoding.
	blobErr := ioutil.WriteFile(relativePath("blob"), []byte(hexValue), 0777)
	checkError(blobErr)
	commitErr := ioutil.WriteFile(relativePath("commit.json"), generateCommit(hexValue), 0777)
	checkError(commitErr)
	refFileErr := ioutil.WriteFile(".cap/refs/main", []byte(hexValue), 0777)
	checkError(refFileErr)
}

//Get current commit hash in ref
//Set new commit to point to commit hash in ref
func generateCommit(hexValue string) []byte {
	attributes := map[string]string{"blob": string(hexValue), "previousCommit": "none"}
	commitContent, _ := json.Marshal(attributes)
	return commitContent
}

//Looking at the other ("remote") copy of the repo
//Look at the remote ref
//If behind,
//Check if the commits in the ref are the same, if not,
//go through linked list of commits
//If diverged, then serve an error
//Copy all remote objects into local repo
//Update local ref
//If ahead, do nothing!
func pull() {

}

func push() {

}

func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func relativePath(file string) (filename string) {
	return directory + "/" + file
}