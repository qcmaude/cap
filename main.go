package main

import "os"
import "fmt"
import "log"
import "io/ioutil"
import "github.com/codahale/blake2"

const filename = "file.txt"

var commands = map[string]func(){
	"commit": commit,
	"create": create,
	"pull":   pull,
	"push":   push,
}

func main() {
	commands[os.Args[1]]()
}

func create() {
	err := os.Mkdir(".cap", 0777)
	if err != nil {
		log.Fatal(err)
	}
}

//Record the blob
//Making a hash of the blob
//Make a commit pointing to the blob (json file)
//Update the ref (hash of commit)
func commit() {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	hash = blake2.NewBlake2B()
	hash.Write(bytes)
	sum := hash.Sum(nilf)
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
