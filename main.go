package main

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/codahale/blake2"
)

//What would you recommend is the best way to set these up so as
//to keep everything DRY but also modular?
//My instinct is to make a sort of ".cap directory" struct
//that has a base directory name (.cap), and a list of directories
//(refs, objects, etc).

//Here we have a recursive definition of a directory
type Directory struct {
	name        string
	directories []Directory
	files       []string
}

//We could also use a map-like structure:
// var capDirectories = map[string]string{
// 		"baseDirectory": ".cap",
// 		"objects": ".cap/objects",
// 		"refs": ".cap/refs"
// }

var mainDirectory = Directory{name: ".cap", directories: []Directory{Directory{name: "refs", directories: []Directory{Directory{name: "head", files: []string{"main"}}}}, Directory{name: "objs"}}}

var commands = map[string]func(){
	"commit": commit,
	"create": create,
	"pull":   pull,
	"push":   push,
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("please provide a valid cap command")
	} else {
		command, ok := commands[os.Args[1]]
		if !ok {
			log.Fatal("that is not a valid cap command.")
		}
		command()
	}
}

//1. Create necessary directories for a 'cap' project
//   a. .cap directory
//   b. .cap/refs directory (with /heads, /remotes and later /tags)
//   c. .cap/objects directory (with all commits and blobs)
func create() {
	err := os.MkdirAll(".cap/refs/heads", 0777)
	checkError(err)
	refFile, err := os.Create(".cap/refs/heads/main")
	checkError(err)
	refFile.Close()
	err = os.Mkdir(".cap/objects", 0777)
	checkError(err)
	err = ioutil.WriteFile(".cap/HEAD", []byte("ref/heads/main"), 0666)
	checkError(err)
}

//1. Record the blob (root of project)
//2. Make a hash of the blob
//3. Make a commit (json file) pointing to the blob
//4. Make a new directory for the commit
//4. Update local ref of the current branch
func commit() {
	root, err := saveBlob("file.txt")
	checkError(err)
	//Throw error if there isn't a commit message.
	//TODO: Is this something we want to enforce?
	if len(os.Args) < 3 {
		log.Fatal("please provide a commit message")
	}
	commit, err := saveCommit(root)
	checkError(err)
	err = ioutil.WriteFile(".cap/refs/heads/main", []byte(commit), 0666)
	checkError(err)
}

// Looking at the other ("remote") copy of the repo
// For now, this will be another copy of a 'cap' project
// elsewhere on the same machine.
// 1. Look at the commit in the remote ref.
//    a. Go through linked list of commits to
//       compare local ref to remote ref.
//    b. If refs have diverged, serve an error
//    c. If local ref is ahead, do nothing
// 2. Copy all remote objects into local repo
// 3. Update local ref (if necessary)
func pull() {
	// createConnection()
	// createRemoteRefs() (create .cap/refs/remote)
}

func push() {
	// createConnection()
	// createRemoteRefs()
}

func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

//Uses Blake2 to generate a hash given []byte
func blake2b(bytes []byte) []byte {
	hash := blake2.NewBlake2B()
	hash.Write(bytes)
	sum := hash.Sum(nil)
	return sum
}

//Use this method to create individual file blobs
//TODO: Create a new method (similar to this one) to hash directories
func saveBlob(file string) (string, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	hex := hex.EncodeToString(blake2b(bytes))
	err = ioutil.WriteFile(".cap/objects/"+hex, bytes, 0666)
	if err != nil {
		return "", err
	}
	return hex, nil
}

//1. Read the commit under the local ref for the current branch
//2. Create directory for commit
//3. Create JSON for commit
//4. Hash commit JSON
//5. Create file with hash as title

//TODO: There is no canonical form for json; we're relying on the fact that the json
//package produces consistent output. (We may be able to not keep the serialized bytes
//to verify the hash)
func saveCommit(root string) (string, error) {
	previousCommit, err := readPreviousCommit()
	if err != nil {
		return "", err
	}

	jsonAttributes := map[string]string{"root": root,
		"previous":  previousCommit,
		"message":   os.Args[2],
		"timestamp": time.Now().String()}
	commitContent, _ := json.Marshal(jsonAttributes)
	hash := hex.EncodeToString(blake2b(commitContent))
	err = os.Mkdir(".cap/objects/"+hash[:2], 0777)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(".cap/objects/"+hash+".json", commitContent, 0666)
	if err != nil {
		return "", err
	}

	return hash, nil
}

//Read local ref of current branch
func readPreviousCommit() (string, error) {
	contents, err := ioutil.ReadFile(".cap/refs/heads/main")
	if err != nil {
		return "", err
	}
	return string(contents), nil
}
