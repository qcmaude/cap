package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/codahale/blake2"
)

var commands = map[string]func(){
	"commit": commit,
	"create": create,
	"pull":   pull,
	"push":   push,
	"diff":   diff,
	"clone":  clone,
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
	if len(os.Args) < 3 {
		log.Fatal("please provide a commit message")
	}
	commit, err := saveCommit(root)
	checkError(err)
	err = ioutil.WriteFile(".cap/refs/heads/main", []byte(commit), 0666)
	checkError(err)
}

//Prints out the difference between working directory and last commit
func diff() {
	commit, err := readCurrentCommit()
	var v struct{ Root string }
	err = readJSONFile(".cap/objects/"+commit+".json", &v)
	if err != nil {
		log.Println("cannot read:", err)
		os.Exit(2)
	}

	cmd := exec.Command("diff", "file.txt", ".cap/objects/"+v.Root)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if isExitStatus(err, 1) {
		os.Exit(1)
	}
	if err != nil {
		log.Println("diff:", err)
		os.Exit(2)
	}
}

type remote interface {
	getObjects(hashes []string) (contents [][]byte, err error)
	listObjects() (hashes []string, err error)
	listBranches() (refs [][2]string, err error)
}

func clone() {
	//Give it a remote location (optionally local location)
	// - assuming directory
	//Local needs to init a new repo (create())
	//Takes (.cap/objects) objects in that location and copies to current location
	//Populate local refs/remote/origin
	//Checkout current commit of local repo

	if len(os.Args) < 3 {
		log.Fatal("please provide a remote repo location")
	}
	create()
	rem, err := newRemote(os.Args[2])
	checkError(err)
	err = copyRemoteObjects(rem)
	checkError(err)
	err = os.MkdirAll(".cap/refs/remote/origin", 0777)
	checkError(err)
	refs, err := rem.listBranches()
	checkError(err)
	for _, ref := range refs {
		err = ioutil.WriteFile(".cap/refs/remote/origin/"+ref[0], []byte(ref[1]), 0666)
		checkError(err)
	}

	copyFile(".cap/refs/heads/main", ".cap/refs/remote/origin/main")

	commit, err := readCurrentCommit()
	var v struct{ Root string }
	err = readJSONFile(".cap/objects/"+commit+".json", &v)
	checkError(err)
	err = copyFile("file.txt", ".cap/objects/"+v.Root)
	checkError(err)
}

// copyFile copies the contents of src to dst.
// if dst doesn't exist, it is created with mode 0666,
// otherwise, it is truncated.
func copyFile(dst, src string) error {
	contents, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, contents, 0666)
}

func newRemote(location string) (remote, error) {
	if strings.HasPrefix(location, "/") {
		return &filesystemRemote{dir: location}, nil
	}
	return nil, errors.New("unrecognized location")
}

func copyRemoteObjects(rem remote) error {
	objectNames, err := rem.listObjects()
	if err != nil {
		return err
	}
	objectContents, err := rem.getObjects(objectNames)
	if err != nil {
		return err
	}

	for i, objectName := range objectNames {
		hashedFile := hex.EncodeToString(blake2b(objectContents[i]))
		// Verify if the hashedFile contents match the file title; if not, throw an error
		if hashedFile != objectName && hashedFile+".json" != objectName {
			log.Fatal("remote object name does not match hash")
		}
		err = ioutil.WriteFile(".cap/objects/"+objectName, objectContents[i], 0666)
		if err != nil {
			return err
		}
	}
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
	previousCommit, err := readCurrentCommit()
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
func readCurrentCommit() (string, error) {
	contents, err := ioutil.ReadFile(".cap/refs/heads/main")
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func readJSONFile(filename string, v interface{}) error {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(contents, v)
}
